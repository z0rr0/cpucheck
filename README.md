# CPUcheck

It is an easy CPU check tool.

- There are N workers (where N is runtime.NumCPU).
- A worker gets pseudo random bytes and calculates SHA-256 sum for it.

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

### Parameters

```
Usage of ./cpucheck:
  -s int
        data size (default 65536)
  -t int
        time duration (seconds) (default 10)
```

### Example

```
./cpucheck
Processors      4
Op. system      linux
Architecture    amd64
Data size       65536 bytes
Duration        10 seconds
. . . . . . . . . . .
Results
Worker 1        1086
Worker 4        1073
Worker 2        1085
Worker 3        1084
---
Total   4328
```
