package handlers

import (
	"fmt"
	"os"
	"time"

	w "channels/pkg/webrtc"

	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
)

func Stream(ctx fiber.Ctx) error {
	suuid := ctx.Params("suuid")
	if suuid == "" {
		ctx.Status(400)
		return nil
	}

	ws := "ws"
	if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
		ws = "wss"
	}
	w.RoomsLock.Lock()
	if _, ok := w.Streams[suuid]; ok {
		w.RoomsLock.Unlock()
		return ctx.Render("stream", fiber.Map{
			"StreamWebSocketAddr": fmt.Sprintf("%s://%s/stream/%s/websocket", ws, ctx.Hostname(), suuid),
			"ChatWebSocketAddr":   fmt.Sprintf("%s://%s/stream/%s/chat/websocket", ws, ctx.Hostname(), suuid),
			"ViewerWebSocketAddr": fmt.Sprintf("%s://%s/stream/%s/viewer/websocket", ws, ctx.Hostname(), suuid),
			"Type":                "stream",
		}, "layouts/main")
	}
	w.RoomsLock.Unlock()
	return ctx.Render("stream", fiber.Map{
		"NoStream": "true",
		"Leave":    "true",
	}, "layouts/main")
}

func StreamWebSocket(c *websocket.Conn) {
	suuid := c.Params("suuid")
	if suuid == "" {
		return
	}
	w.RoomsLock.Lock()
	if stream, ok := w.Streams[uuid]; ok {
		w.RoomsLock.Unlock()
		w.StreamConn(c, streams.Peers)
		return
	}
	w.RoomsLock.Unlock()
}

func StreamViewerWebSocket(c *websocket.Conn) {
	suuid := c.Params("suuid")
	if suuid == "" {
		return
	}
	w.RoomsLock.Lock()
	if stream, ok := w.Streams[uuid]; ok {
		w.RoomsLock.Unlock()
		w.ViewerConn(c, streams.Peers)
		return
	}
	w.RoomsLock.Unlock()
}

func ViewerConn(c *websocket.Conn, p *w.Peers) {
	ticker := time.NewTicker(time.Second)
	defer c.Close()
	for {
		select {
		case <-ticker.C:
			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write([]byte(fmt.Sprintf("%d", len(p.Connections))))
		}
	}
}
