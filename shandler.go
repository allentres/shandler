package shandler

import (
	"encoding/json"
	"go.uber.org/zap"
	"io/ioutil"
	"net/http"
	"nhooyr.io/websocket"
	"sync"
)

type HTTPObjects struct {
	Handler *Handler
	Socket  *websocket.Conn
	Writer  http.ResponseWriter
	Request *http.Request
}

type broadcastMessage struct {
	from        ClientHandler
	messageType websocket.MessageType
	data        []byte
}

type ClientHandler interface {
	Connected()
	Disconnected()
	Handle(data []byte)
	HandleJson(message interface{})
	HandleBinary(data []byte)
	Write(data []byte)
	WriteJson(message interface{})
	WriteBinary(data []byte)
}

// ClientSpawner creates a new client handler
type ClientSpawner func(HTTPObjects) ClientHandler

type Handler struct {
	Log           *zap.Logger
	clients       sync.Map
	connect       chan ClientHandler
	disconnect    chan ClientHandler
	broadcast     chan broadcastMessage
	clientSpawner ClientSpawner
}

// NewHandler creates a new handler. Takes a logger and a client spawner function. The spawner will be called for every
// new client connection.
func NewHandler(log *zap.Logger, clientSpawner ClientSpawner) *Handler {
	handler := &Handler{
		Log:           log.Named("shandler"),
		connect:       make(chan ClientHandler),
		disconnect:    make(chan ClientHandler),
		broadcast:     make(chan broadcastMessage),
		clientSpawner: clientSpawner,
	}
	go func(h *Handler) {
		for {
			select {
			case c := <-h.connect:
				h.clients.Store(c, true)
				h.Log.Debug("connect client", zap.Any("clients", h.clients))
			case c := <-h.disconnect:
				h.clients.Delete(c)
				h.Log.Debug("disconnect client", zap.Any("clients", h.clients))
			case message := <-h.broadcast:
				h.clients.Range(func(k, v interface{}) bool {
					c := k.(ClientHandler)
					// don't send to the sender
					if k == message.from {
						return true
					}
					if message.messageType == websocket.MessageBinary {
						c.WriteBinary(message.data)
					} else {
						c.Write(message.data)
					}
					return true
				})
			}
		}
	}(handler)
	return handler
}

func (h *Handler) Handle(w http.ResponseWriter, r *http.Request) {
	c, err := websocket.Accept(w, r, &websocket.AcceptOptions{
		InsecureSkipVerify: true, // disable CORS
	})
	if err != nil {
		h.Log.Error("websocket handler error",
			zap.String("remoteIp", r.RemoteAddr),
			zap.Error(err),
		)
		return
	}
	defer c.Close(websocket.StatusInternalError, "the sky is falling")

	// create a new client
	client := h.clientSpawner(HTTPObjects{h, c, w, r})
	// add it to the handler's clients map
	h.connect <- client
	ctx := r.Context()
	// connected
	client.Connected()
	// process incoming data
	for {
		typ, reader, err := c.Reader(ctx)
		if err != nil {
			break
		}
		data, err := ioutil.ReadAll(reader)
		if err != nil {
			h.Log.Error("error reading from client ws reader", zap.Error(err))
			break
		}
		if typ == websocket.MessageBinary {
			client.HandleBinary(data)
		} else {
			var jsonMessage interface{}
			if json.Unmarshal(data, &jsonMessage) != nil {
				client.Handle(data)
			} else {
				// if this is a ping, pong it
				if jsonMessage, ok := jsonMessage.(map[string]interface{}); ok {
					if re, exists := jsonMessage["re"]; exists && re == "ping" {
						pong, _ := json.Marshal(map[string]interface{}{
							"re": "pong",
						})
						_ = c.Write(ctx, websocket.MessageText, pong)
						continue
					}
				}
				client.HandleJson(jsonMessage)
			}
		}
	}
	h.disconnect <- client
	client.Disconnected()
}

func (h *Handler) Broadcast(from ClientHandler, messageType websocket.MessageType, data []byte) {
	h.broadcast <- broadcastMessage{from, messageType, data}
}

func (h *Handler) BroadcastJson(from ClientHandler, message interface{}) {
	data, err := json.Marshal(message)
	if err != nil {
		h.Log.Error("failed to marshal broadcast json message", zap.Error(err))
		return
	}
	h.Broadcast(from, websocket.MessageText, data)
}
