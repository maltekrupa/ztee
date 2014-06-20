package main

import (
	"net"
	"fmt"
	"io"
)

func WriteIP(ipChan chan net.IP, doneChan chan int) {
	for ip := range ipChan {
		fmt.Println(ip)
	}
	doneChan <- 1
}

func Rewrite(rewriteChan chan string, rewriter io.Writer, doneChan chan int) {
	for line := range rewriteChan {
		io.WriteString(rewriter, line + "\n")
	}
	doneChan <- 1
}