package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"sync"
	"time"
)

type result struct {
	url   string
	count int
	err   error
}

type countHandler func(reader io.Reader) (int, error)

type countRoutine func(ctx context.Context, url string) (int, error)

func main() {
	ctx := context.Background()
	results := make(chan *result)
	reader := strings.NewReader("https://golang.org\nhttps://golang.org\n")
	//reader := os.Stdin
	routine := newCountRoutine(countWord("Go"))
	go scanLines(ctx, reader, routine,5, results)
	resultHandle(ctx, results)
}

func resultHandle(ctx context.Context, results chan *result) {
	total := 0
	for {
		select {
		case r, ok := <-results:
			if !ok {
				fmt.Printf("Total: %v\n", total)
				return
			}
			if r.err != nil {
				fmt.Printf("Error on processing %v: %v\n", r.url, r.err)
				break
			}
			fmt.Printf("Count for %v: %v\n", r.url, r.count)
			total += r.count
		case <-ctx.Done():
			return
		}
	}
}

func scanLines(ctx context.Context, reader io.Reader, routine countRoutine, concurrency int, results chan<- *result) {
	defer close(results)
	scanner := bufio.NewScanner(reader)
	s := newSemaphore(concurrency)
	wg := sync.WaitGroup{}
	for scanner.Scan() {
		text := scanner.Text()
		if err := s.waitOne(ctx); err != nil {
			results <- &result{
				url:   text,
				count: 0,
				err:   err,
			}
		}
		wg.Add(1)
		go func() {
			defer s.release()
			defer wg.Done()
			count, err := routine(ctx, text)
			results <- &result{
				url:   text,
				count: count,
				err:   err,
			}
		}()
	}
	wg.Wait()
}

func newCountRoutine(handler countHandler) countRoutine {
	return func(ctx context.Context, url string) (i int, err error) {
		counter := 0
		if url == "" {
			return counter, fmt.Errorf("url is empty")
		}

		ctx, _ = context.WithTimeout(ctx, 5*time.Second)
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
		if err != nil {
			return counter, fmt.Errorf("error on creating request: %w", err)
		}

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return counter, fmt.Errorf("error on executing request: %w", err)
		}
		defer resp.Body.Close()

		return handler(resp.Body)
	}
}

func countSubstr(substr string) countHandler {
	return func(reader io.Reader) (i int, err error) {
		bytes, err := ioutil.ReadAll(reader)
		if err != nil {
			return 0, err
		}
		return strings.Count(string(bytes), substr), nil
	}
}

func countWord(word string) countHandler {
	return func(reader io.Reader) (int, error) {
		counter := 0
		scanner := bufio.NewScanner(reader)
		scanner.Split(bufio.ScanWords)
		for scanner.Scan() {
			t := scanner.Text()
			if t == word {
				counter++
			}
		}
		return counter, scanner.Err()
	}
}
