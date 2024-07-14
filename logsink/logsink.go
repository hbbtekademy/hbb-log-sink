package logsink

import (
	"bufio"
	"fmt"
	"io"
	"io/fs"
	"os"
	"time"
)

type Params struct {
	LogFile      string
	LogFilePerm  fs.FileMode
	MaxSizeBytes int64
	BufSize      int64
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

	poolSize := 100
	bufPool := make(chan []byte, poolSize)
	for i := 0; i < poolSize; i++ {
		buf := make([]byte, 1024*4)
		bufPool <- buf
	}

	reader := bufio.NewReader(rd)
	for {
		buf := <-bufPool
		n, err := reader.Read(buf)
		if n > 0 {
			_, err = f.Write(buf[:n])
			if err != nil {
				return fmt.Errorf("failed writing buffer to file. error: %v", err)
			}
			fsize += int64(n)
			lines++
			bufPool <- buf
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
			err = os.Rename(params.LogFile, fmt.Sprintf("%s.%d", params.LogFile, time.Now().UnixMilli()))
			if err != nil {
				return fmt.Errorf("failed renaming %s. error: %v", params.LogFile, err)
			}

			f, err = os.OpenFile(params.LogFile, os.O_APPEND|os.O_CREATE|os.O_WRONLY, params.LogFilePerm)
			if err != nil {
				return fmt.Errorf("failed opening log file: %s. error: %v", params.LogFile, err)
			}
		}
	}

	return nil
}
