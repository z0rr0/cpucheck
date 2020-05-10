package main

import (
	"bytes"
	"math/rand"
	"testing"
	"time"
)

func TestGenerate(t *testing.T) {
	const n = 1024
	source := rand.NewSource(int64(time.Nanosecond))
	data1 := Generate(source, n, n)
	data2 := Generate(source, n, n)
	if n1, n2 := len(data1), len(data2); n1 != n2 {
		t.Fatalf("failed data length: %d != %d\n", n1, n2)
	}
	same := true
	for i := range data1 {
		same = same && (data1[i] == data2[i])
	}
	if same {
		t.Error("generated same data, ok?")
	}
}

func TestShowResults(t *testing.T) {
	var (
		b       bytes.Buffer
		total   = []uint{1000, 2000, 3000, 4000, 5000, 6000}
		timeout = 5
	)
	err := ShowResults(total, timeout, &b)
	if err != nil {
		t.Fatal(err)
	}
	expected := "\nResults\nWorker 1\t1000\nWorker 2\t2000\nWorker 3\t3000\n" +
		"Worker 4\t4000\nWorker 5\t5000\nWorker 6\t6000\n---\nTotal\t\t\t21000\n" +
		"Avg per second\t\t4200\nAvg per processor\t3500\nAvg per proc/second\t700\n"
	if s := b.String(); s != expected {
		t.Errorf("failed result string\n%#v", s)
	}
}

func TestWork(t *testing.T) {
	data := []byte{1, 2, 3}
	for name, handler := range algorithms {
		t.Logf("check %v\n", name)
		sourceCh := make(chan []byte)
		resultCh := make(chan int)
		done := make(chan struct{})

		w := Worker{ID: 1, In: sourceCh, Out: resultCh, Done: done, Handler: handler}
		go Work(w)

		sourceCh <- data
		if i := <-resultCh; i != w.ID {
			t.Errorf("failed result [%v] %v != %v\n", name, i, w.ID)
		}
		close(sourceCh)
		<-done
		close(resultCh)
	}
}

func BenchmarkGenerate(b *testing.B) {
	source := rand.NewSource(int64(time.Nanosecond))
	for n := 0; n < b.N; n++ {
		data := Generate(source, defaultDataSize, defaultDataSize)
		if m := len(data); m != defaultDataSize {
			b.Errorf("failed length %d", m)
		}
	}
}
