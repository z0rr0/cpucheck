package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"runtime"
	"time"
)

const (
	// iterations of handlers per call
	iterationsSHA256 = 10
	iterationsMD5    = 60
	iterationsGZIP   = 3
	// changeData is data size difference
	changeData = 100
)

var (
	algorithms = map[string]func(data []byte){
		"sha256": ProcessSHA256,
		"md5":    ProcessMD5,
		"gzip":   ProcessGZIP,
	}
)

// Worker is work data item.
type Worker struct {
	ID      int
	In      <-chan []byte
	Out     chan<- int
	Done    chan struct{}
	Handler func(data []byte)
}

// ProcessSHA256 is SHA-256 CPU load process.
func ProcessSHA256(data []byte) {
	var idx, b int
	n := len(data)
	for i := 0; i < iterationsSHA256; i++ {
		sha256.Sum256(data)
		idx, b = rand.Intn(n), rand.Intn(256)
		data[idx] = byte(b)
	}
}

// ProcessMD5 is MD5 CPU load process.
func ProcessMD5(data []byte) {
	var idx, b int
	n := len(data)
	for i := 0; i < iterationsMD5; i++ {
		md5.Sum(data)
		idx, b = rand.Intn(n), rand.Intn(256)
		data[idx] = byte(b)
	}
}

// ProcessGZIP is gzip compression CPU load process.
func ProcessGZIP(data []byte) {
	var (
		idx, b int
		buf    bytes.Buffer
	)
	zw := gzip.NewWriter(&buf)
	n := len(data)
	for i := 0; i < iterationsGZIP; i++ {
		if _, err := zw.Write(data); err != nil {
			fmt.Printf("failed gzip write: %v\n", err)
			return
		}
		idx, b = rand.Intn(n), rand.Intn(256)
		data[idx] = byte(b)
	}
	if err := zw.Close(); err != nil {
		fmt.Printf("failed gzip proccess closing: %v\n", err)
	}
}

// Work is CPU process handler.
func Work(w Worker) {
	for data := range w.In {
		w.Handler(data)
		w.Out <- w.ID
	}
	close(w.Done)
}

// Generate returns pseudo random bytes.
func Generate(s rand.Source, min, max int) []byte {
	r := rand.New(s)
	k := r.Intn(max - min)
	b := make([]byte, min+k)
	r.Read(b)
	return b
}

func main() {
	var totalCounter uint
	size := flag.Int("s", 65536, "data size (bytes)")
	timeout := flag.Int("t", 10, "time duration (seconds)")
	algorithm := flag.String("a", "sha256", "algorithm (sha256, md5, gzip)")
	flag.Parse()

	handler, ok := algorithms[*algorithm]
	if !ok {
		fmt.Printf("ERROR: unknown algorithm \"%v\"\n", *algorithm)
		os.Exit(1)
	}
	numProc := runtime.NumCPU()
	fmt.Printf("Processors\t%d\n", numProc)
	fmt.Printf("Op. system\t%s\n", runtime.GOOS)
	fmt.Printf("Architecture\t%s\n", runtime.GOARCH)
	fmt.Printf("Algorithm\t%s\n", *algorithm)
	fmt.Printf("Data size\t%d bytes\n", *size)
	fmt.Printf("Duration\t%d seconds\n.", *timeout)

	maxBytes := *size + changeData
	source := rand.NewSource(int64(time.Nanosecond))

	sourceCh := make(chan []byte)
	resultCh := make(chan int)
	done := make([]chan struct{}, numProc)
	// run workers
	for i := 0; i < numProc; i++ {
		done[i] = make(chan struct{})
		w := Worker{ID: i, In: sourceCh, Out: resultCh, Done: done[i], Handler: handler}
		go Work(w)
	}
	period := time.Second * time.Duration(*timeout)
	ticker := time.NewTicker(period)
	secTicker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
		secTicker.Stop()
	}()
	// send tasks to workers
	go func() {
		for {
			select {
			case <-ticker.C: // wait timeout
				close(sourceCh)
				return
			case <-secTicker.C: // show second dot
				fmt.Printf(" .")
			default:
				sourceCh <- Generate(source, *size, maxBytes)
			}
		}
	}()
	total := make([]uint, numProc)
	// aggregate workers results
	go func() {
		for i := range resultCh {
			total[i]++
		}
	}()
	// wait all processes finish
	for i := range done {
		<-done[i]
	}
	close(resultCh)
	// show result
	fmt.Println("\nResults")
	for k, v := range total {
		fmt.Printf("Worker %d\t%d\n", k+1, v)
		totalCounter += v
	}
	fmt.Printf("---\nTotal\t%d\n", totalCounter)
}
