package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"time"
)

type logSink struct {
	maxSizeMB    int64
	maxSizeBytes int64
	logFile      string
	bufSize      int64
}

func main() {
	logSink := parseCmdlineArgs()

	f, err := os.OpenFile(logSink.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Fatalf("failed opening log file: %s. error: %v", logSink.logFile, err)
	}
	defer f.Close()

	fStat, err := f.Stat()
	if err != nil {
		log.Fatalf("failed getting log file stat. error: %v", err)
	}
	fsize := fStat.Size()

	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, logSink.bufSize)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			} else {
				fmt.Printf("failed reading buffer. error: %v\n", err)
				continue
			}
		}
		fsize += int64(n)
		f.Write(buf[:n])
		if fsize >= logSink.maxSizeBytes {
			fsize = 0
			f.Close()
			os.Rename(logSink.logFile, fmt.Sprintf("%s.%d", logSink.logFile, time.Now().UnixMilli()))
			f, err = os.OpenFile(logSink.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
			if err != nil {
				fmt.Printf("failed opening log file: %s. error: %v\n", logSink.logFile, err)
			}
		}

	}
}

func parseCmdlineArgs() *logSink {
	logFile := flag.String("logfile", fmt.Sprintf("./pid-%d.log", os.Getpid()), "full log file path including filename")
	maxSizeMB := flag.Int64("size", 10, "max log file size in MB")
	bufSize := flag.Int64("bufsize", 4*1024, "max read buffer size")

	flag.Parse()

	return &logSink{maxSizeMB: *maxSizeMB, maxSizeBytes: *maxSizeMB * 1024 * 1024, logFile: *logFile, bufSize: *bufSize}
}
