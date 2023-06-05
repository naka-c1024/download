package main

import (
	"download"
	"fmt"
	"os"
)

func main() {
	// 分割数
	const divNum = 5

	argc := len(os.Args)
	if argc == 1 {
		fmt.Fprintln(os.Stderr, "error: empty argument")
		os.Exit(1)
	} else if argc != 2 {
		fmt.Fprintln(os.Stderr, "error: multiple arguments")
		os.Exit(1)
	}

	url := os.Args[1]

	err := download.Do(url, divNum)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Download failed: %s\n", err.Error())
		os.Exit(1)
	}
}
