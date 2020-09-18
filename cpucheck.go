package main

import (
	"bytes"
	"compress/gzip"
	"crypto/md5"
	"crypto/sha256"
	"flag"
	"fmt"
	"io"
	"math"
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
	// changeData is init data size difference
	changeData = 16
	// defaultDataSize is default bytes generation
	defaultDataSize = 65536
)

var (
	algorithms = map[string]func(data []byte){
		"sha256": ProcessSHA256,
		"md5":    ProcessMD5,
		"gzip":   ProcessGZIP,
		"test":   processTest, // test handler
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

// mixData modifies processing data by random bytes.
func mixData(data []byte, m int) {
	var idx, b int
	n := len(data)
	for i := 0; i < m; i++ {
		idx, b = rand.Intn(n), rand.Intn(256)
		data[idx] = byte(b)
	}
}

// Printf is fmt.Fprintf wrapper to don't check errors after each call.
func Printf(err error, w io.Writer, format string, a ...interface{}) error {
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(w, format, a...)
	return err
}

// processTest is dummy handler only for test running.
func processTest(_ []byte) {
	time.Sleep(time.Millisecond * 1500)
}

// ProcessSHA256 is SHA-256 CPU load process.
func ProcessSHA256(data []byte) {
	for i := 0; i < iterationsSHA256; i++ {
		sha256.Sum256(data)
		mixData(data, 1)
	}
}

// ProcessMD5 is MD5 CPU load process.
func ProcessMD5(data []byte) {
	for i := 0; i < iterationsMD5; i++ {
		md5.Sum(data)
		mixData(data, 1)
	}
}

// ProcessGZIP is gzip compression CPU load process.
func ProcessGZIP(data []byte) {
	var buf bytes.Buffer
	zw := gzip.NewWriter(&buf)
	for i := 0; i < iterationsGZIP; i++ {
		if _, err := zw.Write(data); err != nil {
			fmt.Printf("failed gzip write: %v\n", err)
			return
		}
		mixData(data, 10)
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
	var k int
	r := rand.New(s)
	diff := max - min
	if diff > 0 {
		k = r.Intn(max - min)
	}
	b := make([]byte, min+k)
	r.Read(b)
	return b
}

// ShowResults calculates and print results.
func ShowResults(total []uint, timeout int, w io.Writer) error {
	var (
		err          error
		numProc      float64
		totalCounter uint
	)
	_, err = fmt.Fprintln(w, "\nResults")
	if err != nil {
		return err
	}
	for k, v := range total {
		err = Printf(err, w, "Worker %d\t%d\n", k+1, v)
		totalCounter += v
		numProc++
	}
	if err != nil {
		return err
	}
	avgProc := float64(totalCounter) / numProc
	avgSecond := float64(totalCounter) / float64(timeout)
	avgProcSecond := avgSecond / numProc
	err = Printf(err, w, "---\nTotal\t\t\t%d\n", totalCounter)
	err = Printf(err, w, "Avg per second\t\t%v\n", math.Round(avgSecond))
	err = Printf(err, w, "Avg per processor\t%v\n", math.Round(avgProc))
	err = Printf(err, w, "Avg per proc/second\t%v\n", math.Round(avgProcSecond))
	return err
}

// Run does algorithm processing.
func Run(size, timeout, numProc int, algorithm string, w io.Writer) error {
	var err error
	handler, ok := algorithms[algorithm]
	if !ok {
		return fmt.Errorf("unknown algorithm \"%v\"", algorithm)
	}
	err = Printf(err, w, "\nProcessors\t%d\n", numProc)
	err = Printf(err, w, "Op. system\t%s\n", runtime.GOOS)
	err = Printf(err, w, "Architecture\t%s\n", runtime.GOARCH)
	err = Printf(err, w, "Algorithm\t%s\n", algorithm)
	err = Printf(err, w, "Data size\t%d bytes\n", size)
	err = Printf(err, w, "Duration\t%d seconds\n.", timeout)
	if err != nil {
		return err
	}
	maxBytes := size + changeData
	source := rand.NewSource(int64(time.Nanosecond))

	sourceCh := make(chan []byte)
	resultCh := make(chan int)
	resultDone := make(chan struct{})
	done := make([]chan struct{}, numProc)
	// run workers
	for i := 0; i < numProc; i++ {
		done[i] = make(chan struct{})
		w := Worker{ID: i, In: sourceCh, Out: resultCh, Done: done[i], Handler: handler}
		go Work(w)
	}
	period := time.Second * time.Duration(timeout)
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
				sourceCh <- Generate(source, size, maxBytes)
			}
		}
	}()
	total := make([]uint, numProc)
	// aggregate workers results
	go func() {
		for i := range resultCh {
			total[i]++
		}
		close(resultDone)
	}()
	// wait all processes finish
	for i := range done {
		<-done[i]
	}
	close(resultCh)
	// wait all totals count
	<-resultDone
	return ShowResults(total, timeout, w)
}

func main() {
	var knownAlgorithms = []string{"sha256", "md5", "gzip"}

	size := flag.Int("s", defaultDataSize, "data size (bytes)")
	timeout := flag.Int("t", 10, "time duration (seconds)")
	algorithm := flag.String("a", "sha256", "algorithm (sha256, md5, gzip, all)")
	flag.Parse()

	if *timeout < 1 {
		fmt.Printf("ERROR: timeout must be positive, but value is %v\n", *timeout)
		os.Exit(2)
	}
	if *algorithm != "all" {
		// use only one value
		knownAlgorithms = []string{*algorithm}
	}
	numProc := runtime.NumCPU()
	for _, a := range knownAlgorithms {
		err := Run(*size, *timeout, numProc, a, os.Stdout)
		if err != nil {
			fmt.Printf("ERROR: %v\n", err)
			os.Exit(2)
		}
	}
}
