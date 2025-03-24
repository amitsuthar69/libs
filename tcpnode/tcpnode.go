// tcpnode is so far my theoretical assumption of a main routine having both server and client as sub-routines,
// making it act as an independent peer in a p2p environment.
package tcpnode

import (
	"log"
	"net"
	"sync"
	"time"
)

type Peer struct {
	peerId     string // peerId is nodeId
	serverAddr string
	conn       net.Conn
}

// A TCPNode should act as an independent peer in a p2p network
type TCPNode struct {
	nodeId     string
	serverAddr string
	peers      map[string]*Peer
	mu         sync.Mutex
	msgChan    chan Message
	stopChan   chan struct{}
	stopOnce   sync.Once
}

func NewTCPNode(nodeId, serverAddr string) *TCPNode {
	return &TCPNode{
		nodeId:     nodeId,
		serverAddr: serverAddr,
		peers:      make(map[string]*Peer, 0),
		msgChan:    make(chan Message),
		stopChan:   make(chan struct{}),
	}
}

// starts the Node
func (tn *TCPNode) Start() {
	go tn.handleServer()
	go tn.handleClient()
	go tn.ping()
}

// stops the Node and perform clean-ups
func (tn *TCPNode) Stop() {
	tn.stopOnce.Do(func() {
		close(tn.stopChan)
		close(tn.msgChan)
		tn.mu.Lock()
		defer tn.mu.Unlock()
		for peerId, peer := range tn.peers {
			peer.conn.Close()
			delete(tn.peers, peerId)
		}
	})
}

// handleServer creates a tcp listener, accepts client connections and reads from the connection.
func (tn *TCPNode) handleServer() {
	ln, err := net.Listen("tcp", tn.serverAddr)
	if err != nil {
		tn.Stop()
		log.Printf("ERROR: unable to start server: %v\n", err)
	}
	log.Printf("INFO: server listening on port: %s", tn.serverAddr)

	go func() {
		<-tn.stopChan
		ln.Close()
	}()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("ERROR: unable to accept connection: %v\n", err)
			continue
		}
		log.Printf("INFO: accepted connection: %s", conn.RemoteAddr())

		go tn.Read(conn)
	}
}

// Read reads from the given connection and classifies the message into HandShakeMessage, ClientMessage, PingMessage or a Regular DataMessage.
//
// for HandShakeMessage, it sends the message to the msgChan so our client can Dial to it.
func (tn *TCPNode) Read(conn net.Conn) {
	defer conn.Close()

	buff := make([]byte, 1024)
	for {
		n, err := conn.Read(buff)
		if err != nil {
			log.Printf("ERROR: read from %s failed: %v\n", conn.RemoteAddr(), err)
			return
		}

		message := ParseMessage(buff[:n])
		switch message.Type {
		case PingMessage:
			continue
		case HandshakeMessage:
			tn.msgChan <- Message{Payload: message.Payload, Type: ClientMessage, SenderId: message.SenderId}
		case DataMessage:
			func(message Message) {
				log.Printf("INFO: Processed message: %v %v", string(message.SenderId), string(message.Payload))
			}(message)
		}
	}
}

// handleClient reads from the msgChan, if received a HandShakeMessage, it Dials to the provided serverAddr in payload.
func (tn *TCPNode) handleClient() {
	for {
		select {
		case message := <-tn.msgChan:
			if message.Type == ClientMessage {
				if _, exists := tn.peers[message.SenderId]; !exists {
					conn, err := net.Dial("tcp", string(message.Payload))
					if err != nil {
						log.Printf("ERROR: unable to dial: %v, %v\n", string(message.Payload), err)
					}
					tcpConn, ok := conn.(*net.TCPConn)
					if ok {
						// socket level keep-alive just to enure connectivity.
						tcpConn.SetKeepAlive(true)
						tcpConn.SetKeepAlivePeriod(30 * time.Second)
					}
					if _, err = conn.Write([]byte("HS:" + tn.nodeId + ":" + tn.serverAddr)); err != nil {
						log.Printf("ERROR: unable to write HS: %v, %v\n", string(message.Payload), err)
						conn.Close()
						continue
					}
					tn.mu.Lock()
					tn.peers[message.SenderId] = &Peer{
						peerId:     tn.nodeId,
						serverAddr: string(message.Payload),
						conn:       conn,
					}
					tn.mu.Unlock()
				}
			}
			if message.Type == DataMessage {
				log.Printf("INFO: Client Processed message: %v %v", string(message.SenderId), string(message.Payload))
				// we will decide what to do with data later...
			}
		case <-tn.stopChan:
			return
		}
	}
}

// pings all the peer connections
func (tn *TCPNode) ping() {
	ticker := time.NewTicker(2 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			for peerId, peer := range tn.peers {
				if _, err := peer.conn.Write([]byte("PING")); err != nil {
					tn.mu.Lock()
					log.Printf("ERROR: ping failed for peer %s: %v", peerId, err)
					peer.conn.Close()
					delete(tn.peers, peerId)
					tn.mu.Unlock()
					return
				}
			}
		case <-tn.stopChan:
			return
		}
	}
}
