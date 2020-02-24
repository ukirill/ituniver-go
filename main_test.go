package main

import (
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountSubstr(t *testing.T) {
	cases := []struct {
		name   string
		source string
		aim    string
		count  int
	}{
		{
			name:   "Five Go",
			source: "GoGOgo Go Go gO GoGo",
			aim:    "Go",
			count:  5,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			counterFunc := countSubstr(c.aim)
			reader := strings.NewReader(c.source)
			i, err := counterFunc(reader)
			assert.Nil(t, err)
			assert.Equal(t, c.count, i)
		})
	}
}

func TestCountWord(t *testing.T) {
	cases := []struct {
		name   string
		source string
		aim    string
		count  int
	}{
		{
			name:   "Five Go",
			source: "GoGOgo Go Go gO GoGo",
			aim:    "Go",
			count:  2,
		},
	}
	for _, c := range cases {
		c := c
		t.Run(c.name, func(t *testing.T) {
			counterFunc := countWord(c.aim)
			reader := strings.NewReader(c.source)
			i, err := counterFunc(reader)
			assert.Nil(t, err)
			assert.Equal(t, c.count, i)
		})
	}
}

func TestScanLines(t *testing.T) {
	cases := map[string]struct{
		count int
		err error
	}{
		"one": {1, nil},
		"two": {2, nil},
		"three": {3, nil},
	}
	readerSource := "one\ntwo\nthree\n"

	t.Run("Lines processing in one routine", func(t *testing.T) {
		ctx := context.Background()
		reader := strings.NewReader(readerSource)
		routine := func(ctx context.Context, url string) (int, error) {
			res := cases[url]
			return res.count, nil
		}
		results := make(chan *result, 3)

		scanLines(ctx, reader, routine, 1, results)
		for r := range results {
			assert.Equal(t, r.count, cases[r.url].count)
			assert.Equal(t, r.err, cases[r.url].err)
		}
	})

	t.Run("Lines processing concurrently", func(t *testing.T) {
		ctx := context.Background()
		reader := strings.NewReader(readerSource)
		routine := func(ctx context.Context, url string) (int, error) {
			res := cases[url]
			return res.count, nil
		}
		results := make(chan *result, 3)

		scanLines(ctx, reader, routine, 3, results)
		for r := range results {
			assert.Equal(t, r.count, cases[r.url].count)
			assert.Equal(t, r.err, cases[r.url].err)
		}
	})
}

func TestCountRoutine(t *testing.T) {
	payload := "GoGoGo"
	wantedCount := 3
	ctx := context.Background()
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
		w.Write([]byte(payload))
	})
	server := httptest.NewServer(handler)
	defer server.Close()

	counterHandler := func(reader io.Reader) (int, error) {
		bytes, err := ioutil.ReadAll(reader)
		assert.Nil(t, err)
		assert.Equal(t, payload, string(bytes))
		return wantedCount, nil
	}

	routine := newCountRoutine(counterHandler)
	t.Run("Test http request routine", func(t *testing.T) {
		count, err := routine(ctx, server.URL)
		assert.Equal(t, wantedCount, count)
		assert.Nil(t, err)
	})
}






