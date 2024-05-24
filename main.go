package main

import (
	"bufio"
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
	file := flag.String("f", "", "list of urls")
	reqs := flag.Int("n", 0, "number of requests")

	flag.Parse()

	if *reqs == 0 {
		fmt.Println(color.RedString("number of requests is not passed"))
		flag.Usage()
		return
	}

	if *url == "" && *file == "" {
		fmt.Println(color.RedString("either url or file should be passed"))
		flag.Usage()
		return
	}

	if *url != "" && *file != "" {
		fmt.Println(color.RedString("either url or file should be passed, not both"))
		flag.Usage()
		return
	}

	f := flag.Lookup("f")

	if f.Value.String() != "" {
		file_content, err := os.Open(*file)
		if err != nil {
			log.Fatal(err)
		}

		scanner := bufio.NewScanner(file_content)

		for scanner.Scan() {
			line := scanner.Text()
			fmt.Println(line)
			var wg sync.WaitGroup
			wg.Add(1)

			go func() {
				code, err := requestStatus(line)

				if err != nil {
					log.Fatal(err)
				}
				fmt.Printf("status code: %d\n", code)
				sendRequest(line, *reqs)

				wg.Done()
			}()

			wg.Wait()
		}

		if scanner.Err() != nil {
			log.Println(scanner.Err())
		}
	} else {
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
		start := time.Now()
		req, err := http.NewRequest("GET", url, nil)
		end := time.Now()

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

		fmt.Println("time for the request: ", end.Sub(start))
		success++
	}

	var final int
	for _, value := range avg_processing {
		final += value
	}
	avg := final / len(avg_processing)
	fmt.Println("final average server processing time: ", color.CyanString(strconv.Itoa(avg))+color.CyanString("ms"))

	fmt.Println("Results:")
	fmt.Printf(color.GreenString("Success ..................................: %d\n"), success)
	fmt.Printf(color.RedString("Failed  ..................................: %d\n"), failure)

	return success, failure, nil
}

func calc(d time.Duration) int {
	return int(d / time.Millisecond)
}
