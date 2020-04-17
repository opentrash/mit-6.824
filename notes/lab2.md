# lab2

#### Reasons why we need GFS
1.  Failures could be very normal because there would be thousands of machines and clients running at the same time.
2.  Files are huge by traditional standard.
3.  Most files are mutated by appending new data rather than overwriting existing data.
4.  Co-designing the applications and the file system API benefits the overall system by increasing our flexibility.


## Architecture
*  single **master**
*  multiple **chunkservers**
*  multiple **clients**

    Clients never read and write file data through the master. Instead, a client asks the master which chunkservers it should contact. It caches this information for a limited time and interacts with chunkservers directly for many subsequent operations.

## Chunk Size
64MB, which is much larger than typical file system block sizes.

#### advantages
-  It reduces clients' need to interact with the master
-  A client is more likely to perform many operations on a given chunk, it can reduce network overhead by keeping a persistent TCP connection to the chunkserver over an extended period of time.
-  It reduces the size of the metadata stored on the master.

#### disadvantages
-  A small file consists of a small number of chunks, perhaps just one.

## Metadata
The master stores three major types of metadata.
-  the file and chunk namespaces
-  the mapping from files to chunks
-  the locations of each chunk's replicas

1. All metadata is kept in master's **memory**.
2. The first two types (namespaces and file-to-chunk mapping) are also kept persistent by logging mutations to an operation log stored on the master’s local disk and replicated on remote machines. Using a log allows us to update the master state simply, reliably, and without risking inconsistencies in the event of a master crash.
3. The master does not store chunk location information persistently. Instead, it asks each chunkserver about its chunks at master startup and whenever a chunkserver joins the cluster.

Keeping operation log is a very important strategy to ensure that all chunkservers sync correctly.

## Consistency Model
1. state of a file region after a data mutation depends on
    - the type of mutation
    - whether it succeeds or fails
    - whether there are concurrent mutations
2. after a sequence of successfull mutations
    - applying mutations to a chunk in the same order on all its replicas
    - using chunk version numbers to detect any replica that has become stale

## System Interactions

### Leases and Mutation Order
1. The master grants a chunk lease to one of the replicas, which we call the primary.
2. The primary picks a serial order for all mutations to the chunk. All replicas follow this order when applying mutations.

### Flow between Master and Replicas
1. The client asks the master which chunkserver holds the current lease for the chunk and the locations of the other replicas. **If no one has a lease, the master grants one to a replica it chooses**.
2. The master replies with the identity of the primary and the locations of the other (secondary) replicas. The client caches this data for future mutations. It needs to contact the master again **only** when the primary becomes unreachable or replies that it no longer holds a lease.
3. The client pushes the data to all the replicas. A client can do so in any order. Each chunkserver will store the data in an internal LRU buffer cache until the data is used or aged out. By decoupling the data flow from the control flow, we can improve performance by scheduling the expensive data flow based on the network topology regardless of which chuknserver is the primary.
4. Once all the replicas have acknowledged receiving the data, the client sends a write request to the primary. The request identifies the data pushed earlier to all of the replicas. The primary assigns consecutive serial numbers to all the mutations it receives, possibly from multiple clients, which provides the necessary serialization. It applies the mutation to its own local state in serial number order.
5. The primary forwards the write request to all secondary replicas. Each secondary replica applies mutations in the same serial number order assigned by the primary.
6. The secondaries all reply to the primary indicating that they have completed the operation.
7. The primary replies to the client. Any errors encountered at any of the replicas are reported to the client. **In case of errors**, the write may have succeeded at the primary and an arbitrary subset of the secondary replicas. (If it had failed at the primary, it would not have been assigned a serial number and forward.) The client request is considered to have failed, and the modified region is left in an **inconsistent state**. Our client code handles such errors by retrying the failed mutation. It will make a few attempts at steps (3) through (7) before failing back to a retry from the beginning of the write.
