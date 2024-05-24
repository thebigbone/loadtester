### HTTP load tester
![http](https://github.com/thebigbone/loadtester/assets/95130644/4bfcc6c5-f9ab-4a59-a3aa-46a10ad27c7a)

- stress test API endpoints or websites
- get the precise response time for each request

### Installation

- install `go`
- run `go mod tidy` for installing dependencies
- build it `go build` or directly run it `go run .`

### Usage

```
-f string
    list of urls
-n int
    number of requests
-u string
    url to load test
```

- provide a file containing multiple urls using `-f` (each url in a newline)
- specify a single url using `-u`
- number of requests to process with `-r` (mandatory flag)
