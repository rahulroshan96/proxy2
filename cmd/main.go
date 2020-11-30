package main

import "github.com/rahulroshan96/proxy2/server"

func main() {
	server := server.NewServer()
	server.Run()
}
