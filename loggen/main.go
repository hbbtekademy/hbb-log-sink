package main

import "fmt"

// Execute following commands on Linux to run benchmarks:
//
//	go build -o loggen.exe loggen/main.go
//	go build -o logsink.exe main.go
//	time ./loggen.exe > test1.log
//	time ./loggen.exe | ./logsink.exe -logfile test2.data -max-size-mb 2000
func main() {
	for i := 0; i < 5000000; i++ {
		fmt.Printf("Lorem ipsum dolor sit amet, consectetur adipiscing elit, sed do eiusmod tempor incididunt ut labore et dolore magna aliqua. Ut enim ad minim veniam, quis nostrud exercitation ullamco laboris nisi ut aliquip %d\n", i)
	}
}
