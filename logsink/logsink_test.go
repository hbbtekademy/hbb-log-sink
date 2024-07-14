package logsink

import (
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestLogSink(t *testing.T) {
	tests := []struct {
		name             string
		logDir           string
		setupFunc        func(logDir string) error
		cleanupFunc      func(logDir string) error
		params           *Params
		logs             []string
		expectedLogFiles int
	}{
		{
			name:   "TC1",
			logDir: "../testdata/TC1",
			setupFunc: func(logDir string) error {
				return os.MkdirAll(logDir, 0700)
			},
			cleanupFunc: func(logDir string) error {
				return os.RemoveAll(logDir)
			},
			params: &Params{
				LogFile:      "../testdata/TC1/TC1.log",
				LogFilePerm:  0640,
				MaxSizeBytes: 12,
				BufSize:      10,
			},
			logs:             []string{"Log line1\n", "Log line2\n", "Log line3\n"},
			expectedLogFiles: 2,
		},
		{
			name:   "TC2",
			logDir: "../testdata/TC2",
			setupFunc: func(logDir string) error {
				return os.MkdirAll(logDir, 0700)
			},
			cleanupFunc: func(logDir string) error {
				return os.RemoveAll(logDir)
			},
			params: &Params{
				LogFile:      "../testdata/TC2/TC2.log",
				LogFilePerm:  0640,
				MaxSizeBytes: 12,
				BufSize:      10,
				Compress:     true,
			},
			logs:             []string{"Log line1\n", "Log line2\n", "Log line3\n"},
			expectedLogFiles: 2,
		},
	}

	errChan := make(chan error, 1)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFunc == nil || tc.cleanupFunc == nil {
				t.Fatalf("setupFun and/or cleanupFunc not defined")
			}

			tc.cleanupFunc(tc.logDir)
			err := tc.setupFunc(tc.logDir)
			if err != nil {
				t.Fatalf("failed setup. error: %v", err)
			}

			r, w, err := os.Pipe()
			if err != nil {
				t.Fatalf("failed getting pipes. error: %v", err)
			}

			var wg sync.WaitGroup
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := Run(tc.params, r)
				errChan <- err
			}()

			for _, l := range tc.logs {
				w.Write([]byte(l))
			}

			// Wait for gzip completes
			if tc.params.Compress {
				time.Sleep(1 * time.Second)
			}

			w.Close()
			wg.Wait()

			select {
			case err := <-errChan:
				if err != nil {
					t.Fatalf("logsink failed. error: %v", err)
				}
			default:
			}

			actualLogFiles := getFiles(tc.logDir)
			if len(actualLogFiles) != tc.expectedLogFiles {
				t.Fatalf("expected: %d log files but got: %d", tc.expectedLogFiles, len(actualLogFiles))
			}
			if tc.params.Compress {
				gzipCount := 0
				for _, l := range actualLogFiles {
					if strings.HasSuffix(l, ".gz") {
						gzipCount++
					}
				}
				if gzipCount != len(actualLogFiles)-1 {
					t.Fatalf("expected: %d gzipped files but got: %d", len(actualLogFiles)-1, gzipCount)
				}
			}

			err = tc.cleanupFunc(tc.logDir)
			if err != nil {
				t.Fatalf("failed cleanup. error: %v", err)
			}

		})
	}
}

func getFiles(dir string) []string {
	files := []string{}
	filepath.WalkDir(dir, func(path string, d fs.DirEntry, err error) error {
		if !d.IsDir() {
			files = append(files, path)
		}
		return nil
	})

	return files
}
