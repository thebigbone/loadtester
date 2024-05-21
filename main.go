package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httptrace"
	"os"
	"strconv"
	"sync"
	"time"

	"github.com/fatih/color"
)

func main() {
	url := flag.String("u", "", "url to load test")
	reqs := flag.Int("n", 0, "number of requests")

	flag.Parse()

	if *url == "" || reqs == nil {
		flag.Usage()
		os.Exit(1)
	}

	code, err := requestStatus(*url)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Printf("status code: %d\n", code)
	var wg sync.WaitGroup
	wg.Add(1)

	go func() {
		sendRequest(*url, *reqs)
		wg.Done()
	}()

	wg.Wait()

}

func requestStatus(url string) (int, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
		return 0, err
	}

	code := resp.StatusCode
	return code, nil
}

func sendRequest(url string, reqs int) (int, int, error) {
	success := 0
	failure := 0
	var avg_processing []int

	var t1, t3, t4 time.Time
	for i := 0; i < reqs; i++ {
		req, err := http.NewRequest("GET", url, nil)

		trace := &httptrace.ClientTrace{
			DNSDone: func(_ httptrace.DNSDoneInfo) { t1 = time.Now() },
			ConnectStart: func(_, _ string) {
				if t1.IsZero() {
					t1 = time.Now()
				}
			},
			GotConn:              func(_ httptrace.GotConnInfo) { t3 = time.Now() },
			GotFirstResponseByte: func() { t4 = time.Now() },
		}
		req = req.WithContext(httptrace.WithClientTrace(context.Background(), trace))

		_, err = http.DefaultTransport.RoundTrip(req)

		if err != nil {
			failure++
		}
		avg_processing = append(avg_processing, calc(t4.Sub(t3)))
		fmt.Println(avg_processing)
		fmt.Println(len(avg_processing))

		success++
	}

	var final int
	for _, value := range avg_processing {
		final += value
	}
	avg := final / len(avg_processing)
	fmt.Println("final average server processing time: ", color.CyanString(strconv.Itoa(avg))+color.CyanString("ms"))

	fmt.Printf("success: %d\n", success)
	fmt.Printf("failed: %d\n", failure)

	return success, failure, nil
}

func calc(d time.Duration) int {
	return int(d / time.Millisecond)
}
