package main

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"net"
	"os"
	"strings"
)

var headerTemplate = []string{
	"HTTP/1.1 %v %v",
	"Content-Type: %v",
	"Connection: close",
	"\n",
}

func extractPath(req *bufio.Reader) string {
	line, err := req.ReadSlice('\n')
	if err != nil {
		return ""
	}
	firstline := string(line)
	path := strings.Split(firstline, " ")[1]
	return path
}

func headers(statusCode uint, statusMessage string, mimeType string) []byte {
	template := strings.Join(headerTemplate, "\n")
	return []byte(fmt.Sprintf(template, statusCode, statusMessage, mimeType))
}

func getContentMIMEType(path string) string {
	tokens := strings.Split(path, ".")
	extension := tokens[len(tokens)-1]
	switch extension {
	case "html":
		return "text/html"
	case "js":
		return "text/js"
	case "png":
		return "image/png"
	case "jpg":
		return "image/jpeg"
	default:
		return "text/plain"
	}
}

func loadFile(path string) ([]byte, error) {
	if strings.Contains(path, "..") {
		return nil, errors.New("no hack please")
	}
	base := "./public"
	filepath := base + path
	file, err := os.Open(filepath)
	if err != nil {
		return nil, err
	}
	r := bufio.NewReader(file)
	buffer := make([]byte, 4096*8)
	n, err := r.Read(buffer)
	if err != nil {
		return nil, err
	}
	return buffer[:n], nil
}

func serveHTTP(conn net.Conn) {
	defer conn.Close()
	log.Printf("Serving %v...\n", conn.LocalAddr())
	req := bufio.NewReader(conn)
	path := extractPath(req)
	file, err := loadFile(path)
	var statusCode uint
	var statusMessage string
	var mimeType string
	var body []byte
	if err != nil {
		statusCode = 200
		statusMessage = "OK"
		mimeType = "text/html"
		body, err = loadFile("/error.html")
		if err != nil {
			statusCode = 500
			statusMessage = "Internal Server Error"
			mimeType = "text/plain"
		}
	} else {
		statusCode = 200
		statusMessage = "OK"
		mimeType = getContentMIMEType(path)
		body = file
	}
	headers := headers(statusCode, statusMessage, mimeType)
	response := append(headers, body...)
	conn.Write(response)
}

func main() {
	ln, err := net.Listen("tcp", ":80")
	if err != nil {
		log.Fatal(err)
	}
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatal(err)
		}
		go serveHTTP(conn)
	}
}
