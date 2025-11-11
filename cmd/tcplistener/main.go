package main

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"net"
)

func getLinesChannel(f io.ReadCloser) <-chan string {
	out := make(chan string, 1)

	go func() {
		defer f.Close()
		defer close(out)

		line := ""

		for {
			data := make([]byte, 8)
			c, err := f.Read(data)
			if err == io.EOF {
				break
			}

			data = data[:c]

			if i := bytes.IndexByte(data, '\n'); i != -1 {
				line += string(data[:i])
				out <- line
				line = ""
				data = data[i+1:]
			}

			line += string(data)
		}

		if line != "" {
			out <- line
		}
	}()

	return out
}

func main() {
	l, err := net.Listen("tcp", ":42069")
	if err != nil {
		log.Fatal(err)
	}

	defer l.Close()

	for {
		conn, err := l.Accept()
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("A connection has been accepted")

		lines := getLinesChannel(conn)
		for line := range lines {
			fmt.Printf("%s\n", line)
		}

		fmt.Println("A Connection has been closed")
	}
}
