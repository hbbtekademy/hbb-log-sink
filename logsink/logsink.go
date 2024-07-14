package logsink

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"log"
	"os"
	"os/exec"
	"time"
)

type Params struct {
	LogFile      string
	LogFilePerm  fs.FileMode
	MaxSizeBytes int64
	BufSize      int64
	Compress     bool
}

func Run(params *Params, rd io.Reader) error {
	f, err := os.OpenFile(params.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, params.LogFilePerm)
	if err != nil {
		return fmt.Errorf("failed opening log file: %s. error: %v", params.LogFile, err)
	}
	defer f.Close()

	fStat, err := f.Stat()
	if err != nil {
		return fmt.Errorf("failed getting log file stat. error: %v", err)
	}

	fsize := fStat.Size()
	lines := int64(0)

	buf := make([]byte, params.BufSize)
	reader := bufio.NewReader(rd)
	for {
		n, err := reader.Read(buf)
		if n > 0 {
			_, err = f.Write(buf[:n])
			if err != nil {
				return fmt.Errorf("failed writing buffer to file. error: %v", err)
			}
			fsize += int64(n)
			lines++
		}
		if err != nil {
			if err == io.EOF {
				err := f.Close()
				if err != nil {
					return fmt.Errorf("failed closing %s. error: %v", f.Name(), err)
				}
				break
			}
			return fmt.Errorf("failed reading buffer. error: %v", err)
		}
		if fsize >= params.MaxSizeBytes {
			fsize = 0
			lines = 0

			err := f.Close()
			if err != nil {
				return fmt.Errorf("failed closing %s. error: %v", f.Name(), err)
			}

			rolledOverFile := fmt.Sprintf("%s.%d", params.LogFile, time.Now().UnixMilli())
			err = os.Rename(params.LogFile, rolledOverFile)
			if err != nil {
				return fmt.Errorf("failed renaming %s. error: %v", params.LogFile, err)
			}

			if params.Compress {
				go compress(rolledOverFile)
			}

			f, err = os.OpenFile(params.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, params.LogFilePerm)
			if err != nil {
				return fmt.Errorf("failed opening log file: %s. error: %v", params.LogFile, err)
			}
		}
	}

	return nil
}

func compress(file string) {
	cmd := exec.Command("gzip", file)
	err := cmd.Run()
	if err != nil {
		log.Printf("failed gzipping %s. error: %v", file, err)
	}
}
