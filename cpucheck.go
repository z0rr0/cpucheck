package main

import (
	"crypto/sha256"
	"flag"
	"fmt"
	"math/rand"
	"runtime"
	"time"
)

const (
	// processes per operation
	perOperation = 10
	// changeData is data size difference
	changeData = 100
)

// Worker is work data item.
type Worker struct {
	ID   int
	In   <-chan []byte
	Out  chan<- int
	Done chan<- struct{}
}

// Process is main CPU load process.
func Process(data []byte, m int) {
	var idx, b int
	n := len(data)
	for i := 0; i < m; i++ {
		sha256.Sum256(data)
		idx, b = rand.Intn(n), rand.Intn(256)
		data[idx] = byte(b)
	}
}

// Work is CPU process handler.
func Work(w Worker, n int) {
	for data := range w.In {
		Process(data, n)
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
	size := flag.Int("s", 65536, "data size")
	timeout := flag.Int("t", 10, "time duration (seconds)")
	flag.Parse()

	numProc := runtime.NumCPU()
	fmt.Printf("Processors\t%d\n", numProc)
	fmt.Printf("Op. system\t%s\n", runtime.GOOS)
	fmt.Printf("Architecture\t%s\n", runtime.GOARCH)
	fmt.Printf("Data size\t%d bytes\n", *size)
	fmt.Printf("Duration\t%d seconds\n", *timeout)

	maxBytes := *size + changeData
	source := rand.NewSource(int64(time.Nanosecond))

	sourceCh := make(chan []byte)
	resultCh := make(chan int)
	done := make([]chan struct{}, numProc)
	// run workers
	for i := 0; i < numProc; i++ {
		done[i] = make(chan struct{})
		w := Worker{ID: i + 1, In: sourceCh, Out: resultCh, Done: done[i]}
		go Work(w, perOperation)
	}
	period := time.Second * time.Duration(*timeout)
	ticker := time.NewTicker(period)
	secTicker := time.NewTicker(time.Second)
	defer func() {
		ticker.Stop()
		secTicker.Stop()
	}()
	fmt.Printf(". ")
	// send tasks to workers
	go func() {
		for {
			select {
			// wait timeout
			case <-ticker.C:
				close(sourceCh)
				return
				// show second dot
			case <-secTicker.C:
				fmt.Printf(". ")
			default:
				sourceCh <- Generate(source, *size, maxBytes)
			}
		}
	}()
	total := make(map[int]uint)
	// get workers results
	go func() {
		for id := range resultCh {
			total[id] += 1
		}
	}()
	// wait processes finish
	for i := range done {
		<-done[i]
	}
	close(resultCh)
	// show result
	fmt.Println("\nResults")
	for k, v := range total {
		fmt.Printf("Worker %d\t%d\n", k, v)
		totalCounter += v
	}
	fmt.Printf("---\nTotal\t%d\n", totalCounter)
}
