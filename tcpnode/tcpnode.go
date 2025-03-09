// tcpnode is so far my theoretical assumption of a main routine having both server and client as sub-routines
// hence, making it act as an independent peer in a p2p environment.
//
// It still need a lot of improvements. It works but with workarounds
package tcpnode

import (
	"log"
	"net"
	"slices"
	"strings"
	"sync"
	"time"
)

type client struct {
	conn net.Conn
	addr string
}

type TCPNode struct {
	serverAddr string
	clients    []client
	mu         sync.Mutex
	MsgChan    chan string
	stopChan   chan struct{}
}

func NewTCPNode(serverAddr string) *TCPNode {
	return &TCPNode{
		serverAddr: serverAddr,
		clients:    make([]client, 0),
		MsgChan:    make(chan string),
		stopChan:   make(chan struct{}),
	}
}

func (n *TCPNode) shakeHands(conn net.Conn) {
	buf := make([]byte, 1024)

	// read thy server addr
	x, err := conn.Read(buf)
	if err != nil {
		log.Printf("Error reading handshake: %v", err)
		return
	}

	// tell our server addr
	_, err = conn.Write([]byte(n.serverAddr))
	if err != nil {
		log.Printf("Error writing handshake: %v", err)
		return
	}

	remoteAddr := strings.TrimSpace(string(buf[:x]))
	n.mu.Lock()
	n.clients = append(n.clients, client{conn: conn, addr: remoteAddr})
	n.mu.Unlock()

	n.MsgChan <- remoteAddr
	log.Printf("Handshake successful with %s (local: %s)", remoteAddr, n.serverAddr)
}

func (n *TCPNode) handleServer() error {
	log.Print("Booting up server...")
	ln, err := net.Listen("tcp", n.serverAddr)
	if err != nil {
		log.Printf("Server failed to start: %v", err)
		return err
	}

	go func() {
		<-n.stopChan
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			select {
			case <-n.stopChan:
				return nil
			default:
				log.Printf("Accept error: %v", err)
				continue
			}
		}

		go func(conn net.Conn) {
			defer conn.Close()

			// first, shake hands
			n.shakeHands(conn)

			// then, continue reading...
			buf := make([]byte, 1024)
			for {
				x, err := conn.Read(buf)
				if err != nil {
					log.Printf("Read error from %s: %v", conn.RemoteAddr(), err)
					return
				}
				log.Printf("Message from %s: %s", conn.RemoteAddr(), string(buf[:x]))
			}
		}(conn)
	}
}

func (n *TCPNode) handleClient() error {
	log.Print("Booting up client...")
	for {
		select {
		case <-n.stopChan:
			return nil
		case remoteServerAddr := <-n.MsgChan:
			remoteServerAddr = strings.TrimSpace(remoteServerAddr)

			if n.isalreadyConnected(remoteServerAddr) {
				continue
			}

			conn, err := net.DialTimeout("tcp", remoteServerAddr, 5*time.Second)
			if err != nil {
				log.Printf("Failed to dial %s: %v", remoteServerAddr, err)
				continue
			}

			_, err = conn.Write([]byte(n.serverAddr))
			if err != nil {
				log.Printf("Error sending handshake: %v", err)
				conn.Close()
				continue
			}

			buf := make([]byte, 1024)
			x, err := conn.Read(buf)
			if err != nil {
				log.Printf("Error reading handshake response: %v", err)
				conn.Close()
				continue
			}

			receivedAddr := strings.TrimSpace(string(buf[:x]))

			n.mu.Lock()
			if !n.isalreadyConnected(receivedAddr) {
				n.clients = append(n.clients, client{conn: conn, addr: receivedAddr})
				go n.handleClientConnection(conn, receivedAddr)
			} else {
				conn.Close()
			}
			n.mu.Unlock()
		}
	}
}

func (n *TCPNode) handleClientConnection(conn net.Conn, addr string) {
	defer conn.Close()
	defer n.removeClient(conn)

	buf := make([]byte, 1024)
	for {
		x, err := conn.Read(buf)
		if err != nil {
			log.Printf("Client read error from %s: %v", addr, err)
			return
		}
		log.Printf("Message from %s: %s", addr, string(buf[:x]))
	}
}

func (n *TCPNode) Broadcast(message []byte) {
	log.Print("Broadcasting: ", string(message), " to: ", n.clients)
	for _, c := range n.clients {
		log.Print("client found: ", c)
		n, err := c.conn.Write(message)
		log.Printf("written %d bytes: ", n)
		if err != nil {
			log.Printf("Error sending message to %s: %v", c.addr, err)
		}
	}
}

func (n *TCPNode) removeClient(conn net.Conn) {
	n.mu.Lock()
	defer n.mu.Unlock()
	for i, c := range n.clients {
		if c.conn == conn {
			n.clients = slices.Delete(n.clients, i, i+1)
			break
		}
	}
}

func (n *TCPNode) Start() {
	go n.handleServer()
	go n.handleClient()
}

func (n *TCPNode) Stop() {
	close(n.stopChan)
	n.mu.Lock()
	defer n.mu.Unlock()
	for _, c := range n.clients {
		c.conn.Close()
	}
	n.clients = nil
}

func (n *TCPNode) isalreadyConnected(addr string) bool {
	n.mu.Lock()
	defer n.mu.Unlock()

	if addr == n.serverAddr {
		return true
	}

	ac := false
	for _, c := range n.clients {
		if c.addr == addr {
			ac = true
			break
		}
	}

	return ac
}
