# CPUcheck

It is an easy CPU check tool.

- There are N workers (where N is runtime.NumCPU).
- A worker gets pseudo random bytes and calculates SHA-256 sum for it.
- The cycle repeats for every worker during the time period.

### Build

Use common go-way to build:

```
go build .
```

Examples of cross-compile builds:

```
# MS windows
GOOS=windows GOARCH=amd64 go build -o cpucheck_windows.exe .

# MacOS
GOOS=darwin GOARCH=amd64 go build -o cpucheck_macos .
```

### Test

```
go test -race -bench=. -benchmem -cover -v .
```

### Parameters

```
Usage of ./cpucheck:
  -a string
        algorithm (sha256, md5, gzip) (default "sha256")
  -s int
        data size (bytes) (default 65536)
  -t int
        time duration (seconds) (default 10)
```

### Example

```
Processors      4
Op. system      linux
Architecture    amd64
Algorithm       sha256
Data size       65536 bytes
Duration        10 seconds
. . . . . . . . . .
Results
Worker 1        1045
Worker 2        1020
Worker 3        1009
Worker 4        1013
---
Total   4087
```
