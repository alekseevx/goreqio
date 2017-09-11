package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"sync/atomic"
	"time"
)

type programOpts struct {
	goroutineCount int
	servers        []string
}

func using() {
	programName := filepath.Base(os.Args[0])
	fmt.Printf("%v <goroutineCount> <host:port> <host:port>...\n", programName)
	os.Exit(0)
}

func parseArgs() programOpts {
	args := os.Args[1:]
	if len(args) < 2 {
		using()
	}

	goroutineCount, err := strconv.Atoi(args[0])
	if err != nil {
		using()
	}

	return programOpts{
		goroutineCount: goroutineCount,
		servers:        args[1:]}
}

func doTestTCP(servers []string, reqCounter *uint64, connCount *int32) {
	req := []byte("GET /hello HTTP/1.1\r\n" +
		"Host: 172.17.0.129\r\n" +
		"Content-Type: text/plain\r\n" +
		"Connection: Close\r\n" +
		"\r\n")

	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	readBuf := make([]byte, 2*1024)
	for {
		server := servers[r.Intn(len(servers))]
		conn, err := net.Dial("tcp", server)
		if err != nil {
			fmt.Printf("Can't connect to %v: %v\n", server, err)
			continue
		}
		defer conn.Close()
		atomic.AddInt32(connCount, 1)

		_, err = conn.Write(req)
		if err != nil {
			fmt.Printf("Can't send req %v: %v\n", server, err)
		} else {
			for {
				_, err := conn.Read(readBuf)
				if err != nil {
					if err != io.EOF {
						fmt.Printf("Can't receive resp from %v: %v\n", server, err)
					} else {
						// fmt.Printf("%v\n", string(readBuf))
						atomic.AddUint64(reqCounter, 1)
					}

					break
				}
			}
		}

		atomic.AddInt32(connCount, -1)
	}
}

func doTestHTTP(servers []string, reqCounter *uint64, connCount *int32) {
	s := rand.NewSource(time.Now().Unix())
	r := rand.New(s)

	readBuf := make([]byte, 2*1024)
	client := &http.Client{}

	var reqs []*http.Request
	for i := 0; i < len(servers); i++ {
		req, err := http.NewRequest("GET", "http://"+servers[i]+"/hello", nil)
		req.Header.Add("Host", "172.17.0.129")
		if err != nil {
			fmt.Printf("Can't create HttpRequest: %v\n", err)
			return
		}
		reqs = append(reqs, req)
	}

	for {
		server := servers[r.Intn(len(servers))]
		req := reqs[r.Intn(len(servers))]

		resp, err := client.Do(req)
		if err != nil {
			fmt.Printf("client.Do failed: %v\n", err)
			continue
		}

		defer resp.Body.Close()
		for {
			_, err := resp.Body.Read(readBuf)
			if err != nil {
				if err != io.EOF {
					fmt.Printf("Can't receive resp from %v: %v\n", server, err)
				} else {
					// fmt.Printf("%v\n", string(readBuf))
					atomic.AddUint64(reqCounter, 1)
				}

				break
			}
		}
	}
}

func main() {
	opts := parseArgs()

	rand.Seed(time.Now().Unix())

	var reqCounter uint64
	var connCount int32
	for i := 0; i < opts.goroutineCount; i++ {
		go doTestHTTP(opts.servers, &reqCounter, &connCount)
	}

	startTime := time.Now()
	startReqCounter := reqCounter
	for {
		time.Sleep(2 * time.Second)
		stopReqCounter := reqCounter
		stopTime := time.Now()

		elapsed := stopTime.Sub(startTime)
		fmt.Printf("RPS=%v\n", float64(stopReqCounter-startReqCounter)/elapsed.Seconds())
		fmt.Printf("ConnCount=%v\n", connCount)

		startTime, startReqCounter = stopTime, stopReqCounter
	}
}
