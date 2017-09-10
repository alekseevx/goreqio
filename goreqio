package main

import (
	"fmt"
	"io"
	"math/rand"
	"net"
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

func doTest(servers []string, reqCounter *uint64, connCount *int32) {
	req := []byte("GET /hello HTTP/1.1\r\n")

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
						atomic.AddUint64(reqCounter, 1)
					}

					break
				}
			}
		}

		conn.Close()
		atomic.AddInt32(connCount, -1)
	}
}

func main() {
	opts := parseArgs()

	rand.Seed(time.Now().Unix())

	var reqCounter uint64
	var connCount int32
	for i := 0; i < opts.goroutineCount; i++ {
		go doTest(opts.servers, &reqCounter, &connCount)
	}

	startTime := time.Now()
	startReqCounter := reqCounter
	for {
		time.Sleep(2 * time.Second)
		stopReqCounter := reqCounter
		stopTime := time.Now()

		elapsed := stopTime.Sub(startTime)
		fmt.Printf("RPS=%v", float64(stopReqCounter-startReqCounter)/elapsed.Seconds())
		fmt.Printf("ConnCount=%v", connCount)

		startTime, startReqCounter = stopTime, stopReqCounter
	}
}
