package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"path/filepath"
	"time"
)

type logSinkParams struct {
	logFile      string
	logFilePerm  fs.FileMode
	maxLines     int64
	maxSizeBytes int64
	bufSize      int64
}

func main() {
	params := parseCmdlineArgs()

	f, err := os.OpenFile(params.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, params.logFilePerm)
	if err != nil {
		log.Fatalf("failed opening log file: %s. error: %v", params.logFile, err)
	}
	defer f.Close()

	fStat, err := f.Stat()
	if err != nil {
		log.Fatalf("failed getting log file stat. error: %v", err)
	}

	fsize := fStat.Size()
	lines := int64(0)

	reader := bufio.NewReader(os.Stdin)
	buf := make([]byte, params.bufSize)
	for {
		n, err := reader.Read(buf)
		if err != nil {
			if err == io.EOF {
				break
			}
			log.Printf("failed reading buffer. error: %v", err)
			continue
		}

		_, err = f.Write(buf[:n])
		if err != nil {
			log.Printf("failed writing buffer to file. error: %v", err)
			continue
		}
		fsize += int64(n)
		lines++

		if fsize >= params.maxSizeBytes || (params.maxLines > 0 && lines >= params.maxLines) {
			fsize = 0
			lines = 0

			err := f.Close()
			if err != nil {
				log.Printf("failed closing %s. error: %v", f.Name(), err)
				break
			}
			err = os.Rename(params.logFile, fmt.Sprintf("%s.%d", params.logFile, time.Now().UnixMilli()))
			if err != nil {
				log.Printf("failed renaming %s. error: %v", params.logFile, err)
				break
			}

			f, err = os.OpenFile(params.logFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, params.logFilePerm)
			if err != nil {
				log.Printf("failed opening log file: %s. error: %v", params.logFile, err)
				break
			}
		}
	}
}

func parseCmdlineArgs() *logSinkParams {
	logFile := flag.String("logfile", fmt.Sprintf("./pid-%d.log", os.Getpid()), "Full log file path including filename.")
	logFilePerm := flag.Uint("logfile-perm", 0, "Log file permission. (default 0640)")
	maxSizeMB := flag.Int64("max-size", 10, "Max log file size in MB. If both max-size and max-lines is specified, logfile will rollover when either thresold is reached.")
	maxLines := flag.Int64("max-lines", 0, "Max log file lines.")
	bufSize := flag.Int64("buf-size", 4*1024, "Max read buffer size.")

	flag.Parse()

	if *logFilePerm == 0 {
		*logFilePerm = 0640
	}

	return &logSinkParams{
		logFile:      filepath.Clean(*logFile),
		logFilePerm:  fs.FileMode(*logFilePerm),
		maxSizeBytes: *maxSizeMB * 1024 * 1024,
		maxLines:     *maxLines,
		bufSize:      *bufSize,
	}
}
