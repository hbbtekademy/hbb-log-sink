package logsink

import (
	"io/fs"
	"os"
	"path/filepath"
	"sync"
	"testing"
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
			logDir: "TC1",
			setupFunc: func(logDir string) error {
				return os.Mkdir(logDir, 0700)
			},
			cleanupFunc: func(logDir string) error {
				return os.RemoveAll(logDir)
			},
			params: &Params{
				LogFile:      "TC1/TC1.log",
				LogFilePerm:  0640,
				MaxSizeBytes: 18,
				BufSize:      1024,
			},
			logs:             []string{"Log line1\n", "Log line2\n", "Log line3\n"},
			expectedLogFiles: 2,
		},
	}

	errChan := make(chan error, 1)
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			if tc.setupFunc != nil {
				err := tc.setupFunc(tc.logDir)
				if err != nil {
					t.Fatalf("failed setup. error: %v", err)
				}
			}

			if tc.cleanupFunc != nil {
				defer func() {
					err := tc.cleanupFunc(tc.logDir)
					if err != nil {
						t.Fatalf("failed cleanup. error: %v", err)
					}
				}()
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
