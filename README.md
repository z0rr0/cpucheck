# CPUcheck

It is an easy CPU check tool.

- There are N workers (where N is runtime.GOMAXPROCS).
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
        data size (default 1024)
  -t int
        time duration (seconds) (default 10)
```
