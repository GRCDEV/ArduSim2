package util

import (
	"bufio"
	"io"
	"sync"
)

// LogScanner merges multiple readers (stdout/stderr) into a single scan loop.
type LogScanner struct {
	ch   chan string
	done chan struct{}
	text string
	mu   sync.Mutex
}

func NewLogScanner(readers ...io.Reader) *LogScanner {
	ls := &LogScanner{
		ch:   make(chan string, 100),
		done: make(chan struct{}),
	}

	var wg sync.WaitGroup
	for _, r := range readers {
		if r == nil {
			continue
		}
		wg.Add(1)
		go func(reader io.Reader) {
			defer wg.Done()
			scanner := bufio.NewScanner(reader)
			for scanner.Scan() {
				ls.ch <- scanner.Text()
			}
		}(r)
	}

	go func() {
		wg.Wait()
		close(ls.ch)
		close(ls.done)
	}()

	return ls
}

func (ls *LogScanner) Scan() bool {
	line, ok := <-ls.ch
	if ok {
		ls.mu.Lock()
		ls.text = line
		ls.mu.Unlock()
		return true
	}
	return false
}

func (ls *LogScanner) Text() string {
	ls.mu.Lock()
	defer ls.mu.Unlock()
	return ls.text
}
