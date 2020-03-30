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

All metadata is kept in master's **memory**.

