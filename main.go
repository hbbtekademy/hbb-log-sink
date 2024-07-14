package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"os"
	"path/filepath"

	"github.com/hbbtekademy/hbb-log-sink/logsink"
)

func main() {
	params := parseCmdlineArgs()
	err := logsink.Run(params, os.Stdin)
	if err != nil {
		log.Fatalln(err)
	}
}

func parseCmdlineArgs() *logsink.Params {
	logFile := flag.String("logfile", fmt.Sprintf("./pid-%d.log", os.Getpid()), "Full log file path including filename.")
	logFilePerm := flag.Uint("logfile-perm", 0, "Log file permission. (default 0640)")
	maxSizeMB := flag.Int64("max-size", 10, "Max log file size in MB.")
	bufSize := flag.Int64("buf-size", 4*1024, "Max read buffer size.")

	flag.Parse()

	if *logFilePerm == 0 {
		*logFilePerm = 0640
	}

	return &logsink.Params{
		LogFile:      filepath.Clean(*logFile),
		LogFilePerm:  fs.FileMode(*logFilePerm),
		MaxSizeBytes: *maxSizeMB * 1024 * 1024,
		BufSize:      *bufSize,
	}
}
