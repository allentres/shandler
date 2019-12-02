package main

import (
	"encoding/json"
	"github.com/allentres/shandler"
	"github.com/gorilla/mux"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"net/http"
	"nhooyr.io/websocket"
)

// ----------- Server
func main() {
	cfg := zap.Config{
		Encoding:         "json",
		Level:            zap.NewAtomicLevelAt(zapcore.InfoLevel),
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stdout"},
		EncoderConfig: zapcore.EncoderConfig{
			//TimeKey:        "ts",
			//LevelKey:       "level",
			NameKey: "logger",
			//CallerKey:      "caller",
			MessageKey: "msg",
			//StacktraceKey:  "stacktrace",
			EncodeLevel:    zapcore.CapitalColorLevelEncoder, // zapcore.LowercaseLevelEncoder
			EncodeTime:     zapcore.EpochTimeEncoder,         // zapcore.ISO8601TimeEncoder
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
	}
	logger, err := cfg.Build()
	if err != nil {
		panic(err)
	}
	defer logger.Sync()

	r := mux.NewRouter().StrictSlash(true)
	// ws
	wsHandler := shandler.NewHandler(logger, NewClient)
	r.HandleFunc("/ws", wsHandler.Handle)
	// static files handler
	r.PathPrefix("/").Handler(http.StripPrefix("/", http.FileServer(http.Dir("./ui/dist"))))

	httpServer := http.Server{
		Addr:    "0.0.0.0:8888",
		Handler: r,
	}

	_ = httpServer.ListenAndServe()
	logger.Info("Done")

}

// ---------- Client Handler
// your application-specific implementation goes here.
type Client struct {
	shandler.HTTPObjects
}

// NewClient gets called on every new client connection. This function must return a ClientHandler
func NewClient(httpObjects shandler.HTTPObjects) shandler.ClientHandler {
	return &Client{httpObjects}
}

func (c *Client) Connected() {
	c.Handler.Log.Debug("connected")
}

func (c *Client) Disconnected() {
	c.Handler.Log.Debug("disconnected")
}

func (c *Client) Write(data []byte) {
	c.Handler.Log.Debug("write", zap.ByteString("data", data))
	err := c.Socket.Write(c.Request.Context(), websocket.MessageText, data)
	if err != nil {
		c.Handler.Log.Error("failed to write to websocket", zap.Error(err))
	}
}

func (c *Client) WriteJson(message interface{}) {
	c.Handler.Log.Debug("write json", zap.Any("message", message))
	data, err := json.Marshal(message)
	if err != nil {
		c.Handler.Log.Error("failed to marshal json message", zap.Error(err))
		return
	}
	c.Write(data)
}

func (c *Client) WriteBinary(data []byte) {
	c.Handler.Log.Debug("write binary", zap.ByteString("data", data))
	err := c.Socket.Write(c.Request.Context(), websocket.MessageBinary, data)
	if err != nil {
		c.Handler.Log.Error("failed to write binary to websocket", zap.Error(err))
	}
}

func (c *Client) Handle(data []byte) {
	c.Handler.Log.Debug("handle", zap.ByteString("data", data))
	c.Handler.Broadcast(c, websocket.MessageText, data)
}

func (c *Client) HandleJson(message interface{}) {
	c.Handler.Log.Debug("handle json", zap.Any("message", message))
}

func (c *Client) HandleBinary(data []byte) {
	c.Handler.Log.Debug("handle binary", zap.ByteString("data", data))
}
