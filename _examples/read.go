package main

import (
	"fmt"
	"io"
	"os"

	"github.com/kevin-cantwell/srt"
)

func main() {
	file, err := os.Open(os.Args[1])
	if err != nil {
		panic(err)
	}
	reader := srt.NewReader(file)
	for {
		subtitle, err := reader.ReadSubtitle()
		if err != nil {
			if err == io.EOF {
				os.Exit(0)
			}
			panic(err)
		}
		fmt.Println(subtitle.Number)
		fmt.Println(subtitle.Start, "-->", subtitle.End)
		fmt.Println(subtitle.Text)
		fmt.Println("---------------------")
	}
}
