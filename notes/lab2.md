# lab2

#### Reasons why we need GFS
1. Failures could be very normal because there would be thousands of machines and clients running at the same time.
2. Files are huge by traditional standard.
3. Most files are mutated by appending new data rather than overwriting existing data.
4. Co-designing the applications and the file system API benefits the overall system by increasing our flexibility.


## Architecture
* single **master**
* multiple **chunkservers**
* multiple **clients**

