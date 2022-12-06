package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

const bufferSize int = 25
const drainInterval time.Duration = 10 * time.Second

type RingIntBuffer struct {
	array []int
	pos   int
	size  int
	mx    sync.Mutex
}

func NewRingIntBuffer(size int) *RingIntBuffer {
	return &RingIntBuffer{make([]int, size), -1, size, sync.Mutex{}}
}

func (r *RingIntBuffer) Push(el int) {
	r.mx.Lock()
	defer r.mx.Unlock()
	if r.pos == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			r.array[i-1] = r.array[i]
		}
		r.array[r.pos] = el
	} else {
		r.pos++
		r.array[r.pos] = el
	}
}

func (r *RingIntBuffer) Get() []int {
	if r.pos <= 0 {
		return nil
	}
	r.mx.Lock()
	defer r.mx.Unlock()
	var out []int = r.array[:r.pos+1]
	r.pos = -1
	return out
}

func read(basicChan chan<- int, done chan bool) {
	scan := bufio.NewScanner(os.Stdin)
	var command string
	for scan.Scan() {
		command = scan.Text()
		if strings.EqualFold(command, "exit") {
			fmt.Println("exit from programm")
			close(done)
			return
		}
		i, err := strconv.Atoi(command)
		if err != nil {
			fmt.Println("programm read only integer")
			continue
		}
		basicChan <- i
	}
}

func removeNegative(prevChan <-chan int, nextChan chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-prevChan:
			if data >= 0 {
				nextChan <- data
			}
		case <-done:
			return
		}
	}
}

func removeNotDivTrhree(prevChan <-chan int, nextChan chan<- int, done <-chan bool) {
	for {
		select {
		case data := <-prevChan:
			if data%3 == 0 {
				nextChan <- data
			}
		case <-done:
			return
		}
	}
}

func bufferFunc(prevChan <-chan int, nextChan chan<- int, done <-chan bool, size int, interval time.Duration) {
	buffer := NewRingIntBuffer(size)
	for {
		select {
		case data := <-prevChan:
			buffer.Push(data)
		case <-time.After(interval):
			bufferData := buffer.Get()
			if bufferData != nil {
				for _, data := range bufferData {
					nextChan <- data
				}
			}
		case <-done:
			return
		}
	}
}

func main() {
	input := make(chan int)
	done := make(chan bool)
	go read(input, done)

	negativeFilterChan := make(chan int)
	go removeNegative(input, negativeFilterChan, done)

	ThreeFilterChan := make(chan int)
	go removeNotDivTrhree(negativeFilterChan, ThreeFilterChan, done)

	bufferedChannel := make(chan int)
	go bufferFunc(ThreeFilterChan, bufferedChannel, done, bufferSize, drainInterval)

	for {
		select {
		case data := <-bufferedChannel:
			fmt.Println("data:", data)
		case <-done:
			return
		}
	}
}
