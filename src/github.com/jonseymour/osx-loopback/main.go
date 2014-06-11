package main

import (
	"bufio"
	"encoding/binary"
	"flag"
	"io"
	"log"
	"net"
	"os"
	"time"
)

type ConnectionParams struct {
	addr string
}

type Program struct {
	BurstSize    int64
	BurstDelay   int64
	BurstCount   int64
	InitialDelay int64
	PreambleSize int64
}

func format(p []byte) []byte {
	for i := 0; i < len(p)-1; i++ {
		p[i] = '0' + uint8(i%10)
	}
	p[len(p)-1] = '\n'
	return p
}

func handler(tcpConn *net.TCPConn) {
	var program Program

	var err = binary.Read(bufio.NewReader(tcpConn), binary.LittleEndian, &program)
	if err != nil {
		log.Printf("failed to read program arguments from %s, %s", tcpConn.RemoteAddr(), err)
		return
	}

	log.Printf("program = %+v, addr = %s", program, tcpConn.RemoteAddr())

	var burst = format(make([]byte, program.BurstSize, program.BurstSize))

	if program.PreambleSize > 0 {
		tcpConn.Write(format(make([]byte, program.PreambleSize)))
	}

	time.Sleep(time.Duration(program.InitialDelay) * time.Millisecond)

	if err == nil {
		for i := int64(0); i < program.BurstCount; i++ {
			tcpConn.Write(burst)
			time.Sleep(time.Duration(program.BurstDelay) * time.Millisecond)
		}
	}
	err = tcpConn.Close()
	if err != nil {
		log.Printf("failed to close connection from %s", tcpConn.RemoteAddr())
	}
}

func server(conn ConnectionParams) int {
	var err error
	var reason string
	if tcpaddr, err := net.ResolveTCPAddr("tcp", conn.addr); err == nil {
		if tcpListener, err := net.ListenTCP("tcp", tcpaddr); err == nil {
			log.Printf("listening on %s", conn.addr)
			for true {
				if tcpConn, err := tcpListener.AcceptTCP(); err == nil {
					log.Printf("accepted conntection from %s", tcpConn.RemoteAddr())
					go handler(tcpConn)
				}
			}
		} else {
			reason = "listen failed"
		}
	} else {
		reason = "resolve failed"
	}
	log.Printf("failed to open connection: %s, %v", reason, err)
	return 1
}

func client(conn ConnectionParams, program Program, closeDelay int) int {
	var err error
	if tcpconn, err := net.Dial("tcp", conn.addr); err == nil {
		var br = bufio.NewReader(tcpconn)

		binary.Write(tcpconn, binary.LittleEndian, &program)

		if program.PreambleSize > 0 {
			br.ReadString('\n')
		}

		go func() {
			time.Sleep(time.Duration(program.BurstDelay) * time.Millisecond)
			if tcpconn1, ok := tcpconn.(*net.TCPConn); ok {
				var err1 error
				err1 = tcpconn1.CloseWrite()
				if err1 != nil {
					log.Printf("failed to close connection - %s", err1)
				} else {
					log.Printf("closed write end of connection")
				}
			}
		}()

		var result chan int

		result = make(chan int)

		go func() {
			if written, err := io.Copy(os.Stdout, br); err == nil {
				var expected = program.BurstSize * program.BurstCount
				log.Printf("copied %d bytes of %d expected", written, expected)
				if written != expected {
					result <- 1
				} else {
					result <- 0
				}
			}
		}()

		return <-result
	}
	log.Printf("failed to establish connection: %s", err)
	return 1
}

func main() {
	//
	// The purpose of this program is to demonstrate that when CloseWrite is called on a connection
	// opened across the loopback interface that received (but not acknowledged) packets will be
	// be acknowledged with an incorrect ack sequence number and so prevent the arrival of
	// packets sent after that time.
	//
	// The server waits for connections on the specified interface and port.
	// When it receives a connection it reads parameters from the connection. The parameters are:
	//	   * the number of bytes in each burst
	//    * the number of milliseconds to delay between each burst
	//    * the number of bursts to generate
	//
	// It then enters a loop and generates the specified number of bursts of specified number of characters, with a delay
	// of the specified amount between each burst.
	//
	// The client:
	//	   * connects to the server
	//	   * sends the parameters for the connection to the server
	//	   * creates a goroutine to copy the connections output to stdout and count the response bytes
	//         * delays for a specified number of milliseconds, then issues a CloseWrite on the connection
	//	   * waits for the copying goroutine to finish
	//

	var role string

	var connection ConnectionParams
	var program Program

	var closeDelay int

	flag.StringVar(&role, "role", "client", "The role of this program - either client (default) or server")
	flag.StringVar(&connection.addr, "addr", "127.0.0.1:19622", "The interface")
	flag.Int64Var(&program.BurstSize, "burstSize", 5, "The number of bytes in each burst")
	flag.Int64Var(&program.BurstDelay, "burstDelay", 1000, "The mumber of milliseconds in each burst")
	flag.Int64Var(&program.InitialDelay, "initialDelay", 200, "The mumber of milliseconds to wait before the initial burst")
	flag.Int64Var(&program.BurstCount, "burstCount", 2, "The mumber of bursts to issue before closing the connection")
	flag.Int64Var(&program.PreambleSize, "preambleSize", 69, "The mumber of bytes of preamble to generate on initial response")
	flag.IntVar(&closeDelay, "closeDelay", 0, "The number of milliseconds delay before closing")

	flag.Parse()

	var exit int

	if role == "server" {
		//
		exit = server(connection)

	} else {
		exit = client(connection, program, closeDelay)
	}

	os.Exit(exit)
}
