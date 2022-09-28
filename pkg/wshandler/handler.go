package wshandler

import (
	"context"
	"net/http"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

// WebSocket represents a simplified connection
type WebSocket interface {
	// Wait waits for the connection to close
	Wait()

	// Read reads a message
	Read() (string, bool)

	// Write writes a message
	Write(string) bool
}

type HandleFunc func(WebSocket)

var upgrader websocket.Upgrader

// Handler handles WebSocket connections with a simple messages
type Handler struct {
	conn    *websocket.Conn // underlying connection
	context context.Context // context to cancel the connection
	cancel  context.CancelFunc

	wg sync.WaitGroup // blocks all the ongoing tasks

	// incoming and outgoing tasks
	incoming chan string
	outgoing chan string

	// HandleFunc is called to implement the application logic.
	// It should return to terminate a connection.
	HandleFunc HandleFunc
}

const (
	writeWait      = 10 * time.Second
	pongWait       = time.Minute
	pingInterval   = (pongWait * 9) / 10
	maxMessageSize = 2048
)

func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer h.wg.Wait() // wait for everything to finish before returning to the caller

	var err error
	h.conn, err = upgrader.Upgrade(w, r, nil)
	if err != nil {
		return
	}

	// wait for the context to be cancelled, then close the connection
	h.wg.Add(1)
	go func() {
		defer h.wg.Done()
		<-h.context.Done()
		h.conn.Close()
	}()

	// start receiving and sending messages
	h.wg.Add(2)
	go h.sendMessages()
	go h.recvMessages()

	// start the application logic
	h.wg.Add(1)
	go h.handle()
}

func (h *Handler) handle() {
	defer func() {
		h.wg.Done()
		h.cancel()
	}()

	h.HandleFunc(h)
}

func (h *Handler) sendMessages() {
	// close connection when done!
	defer func() {
		h.wg.Done()
		h.cancel()
	}()

	// setup a timer for pings!
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for {
		select {
		// everything is done!
		case <-h.context.Done():
			return

		// send outgoing messages
		case content := <-h.outgoing:
			if err := h.writeRaw(websocket.TextMessage, []byte(content)); err != nil {
				return
			}
		// send a ping message
		case <-ticker.C:
			if err := h.writeRaw(websocket.PingMessage, []byte{}); err != nil {
				return
			}
		}
	}
}

// writeRaw writes to the underlying socket
func (h *Handler) writeRaw(messageType int, data []byte) error {
	h.conn.SetWriteDeadline(time.Now().Add(writeWait))
	return h.conn.WriteMessage(messageType, data)
}

// Read reads a message sent from the client
func (h *Handler) Read() (message string, ok bool) {
	select {
	case message = <-h.incoming:
		return message, true
	case <-h.context.Done():
		return "", false
	}
}

func (h *Handler) Wait() {
	<-h.context.Done()
}

// Write writes a message to the client
func (sh *Handler) Write(message string) bool {
	select {
	case sh.outgoing <- message:
		return true
	case <-sh.context.Done():
		return false
	}
}

func (h *Handler) recvMessages() {
	// close connection when done!
	defer func() {
		h.wg.Done()
		h.cancel()
	}()

	h.conn.SetReadLimit(maxMessageSize)

	// configure a pong handler
	h.conn.SetReadDeadline(time.Now().Add(pongWait))
	h.conn.SetPongHandler(func(string) error { h.conn.SetReadDeadline(time.Now().Add(pongWait)); return nil })

	// handle incoming messages
	for {
		_, message, err := h.conn.ReadMessage()
		if err != nil {
			return
		}

		// send a message to the incoming channel, or cancel
		select {
		case h.incoming <- string(message):
		case <-h.context.Done():
			return
		}

	}
}

// Reset resets this SocketHandler to default
func (h *Handler) Reset(HandleFunc HandleFunc) {
	h.HandleFunc = HandleFunc

	h.conn = nil

	h.incoming = make(chan string)
	h.outgoing = make(chan string)
	h.context, h.cancel = context.WithCancel(context.Background())
}
