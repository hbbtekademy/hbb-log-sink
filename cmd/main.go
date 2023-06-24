package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"
)

type args struct {
	bufferLines  int
	maxSizeBytes int
	logFilename  string
}

func main() {
	args := parseCmdlineArgs()

	f, err := os.OpenFile(args.logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		fmt.Printf("Failed opening Log File: %s. Error: %v\n", args.logFilename, err)
	}
	defer f.Close()

	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, 4*1024)
	c := 0
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				fmt.Println("break")
				break
			} else {
				fmt.Printf("failed reading buffer. error: %v\n", err)
				continue
			}
		}
		c += n
		f.Write(buf[:n])
		if c >= args.maxSizeBytes {
			c = 0
			f.Close()
			os.Rename(args.logFilename, fmt.Sprintf("%s.%d", args.logFilename, time.Now().UnixMilli()))
			f, err = os.OpenFile(args.logFilename, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("Failed opening Log File: %s. Error: %v\n", args.logFilename, err)
			}
		}

	}
}

func parseCmdlineArgs() *args {
	home := os.Getenv("HOME")
	if strings.TrimSpace(home) == "" {
		home = "."
	}

	args := &args{bufferLines: 10000, maxSizeBytes: 100 * 1024 * 1024, logFilename: fmt.Sprintf("%s\\current.log", home)}

	return args
}
