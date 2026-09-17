package handlers

import (
	"crypto/sha256"
	"fmt"
	"os"
	"time"

	"channels/pkg/chat"
	"channels/pkg/webrtc"
	w "channels/pkg/webrtc"

	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/jcontrib/v3/websocket"
	guuid "github.com/google/uuid"
)

type websocketMessage struct {
	Event string `json:"event"`
	Data  string `json:"data"`
}

func createOrGetRoom(uuid string) (string, string, *w.Room) {
	w.RoomsLock.Lock()
	defer w.RoomsLock.Unlock()
	h := sha256.New()
	h.Write([]byte(uuid))
	suuid := fmt.Sprintf("%x", h.Sum(nil))

	if room := w.Rooms[uuid]; room != nil {
		if _, ok := w.Streams[suuid]; !ok {
			w.Streams[suuid] = room
		}
		return uuid, suuid, room
	}
	p := &w.Peers{
		TrackLocals: make(map[string]*webrtc.TrackLocalStaticRTP),
	}
	room := &w.Room{
		Peers: p,
		Hub:   chat.NewHub(),
	}

	w.Rooms[uuid] = room
	w.Streams[suuid] = room
	go room.Hub.Run()
	return uuid, suuid, room
}

func Room(ctx fiber.Ctx) error {
	uuid := ctx.Params("uuid")
	if uuid == "" {
		ctx.Status(400)
		return nil
	}
	ws := "ws"
	if os.Getenv("ENVIRONMENT") == "PRODUCTION" {
		ws = "wss"
	}

	uuid, suuid, _ := createOrGetRoom(uuid)
	return ctx.Render("peer", fiber.Map{
		"RoomWebsocketAddr":   fmt.Sprintf("%s://%s/room%s/websocket", ws, ctx.Hostname(), uuid),
		"ChatWebSocketAddr":   fmt.Sprintf("%s://%s/room%s/chat/websocket", ws, ctx.Hostname(), uuid),
		"ViewerWebSocketAddr": fmt.Sprintf("%s://%s/room%s/viewer/websocket", ws, ctx.Hostname(), uuid),
		"RoomLink":            fmt.Sprintf("%s://%s/room%s", ctx.Protocol(), ctx.Hostname(), uuid),
		"Streamlink":          fmt.Sprintf("%s://%s/stream%s", ctx.Protocol(), ctx.Hostname(), suuid),
		"Type":                "room",
	}, "layouts/main")
}

func RoomCreate(ctx fiber.Ctx) error {
	return ctx.Redirect().To(fmt.Sprintf("/room/%s", guuid.New().String()))
}

func RoomWebsocket(c *websocket.Conn) {
	uuid := c.Params("uuid")
	if uuid == "" {
		return
	}

	_, _, room := createOrGetRoom(uuid)
	w.RoomConn(c, room.Peers)

	// TODO: Finish
}

func RoomViewerWebsocket(c *websocket.Conn) {
	uuid := c.Params("uuid")
	if uuid == "" {
		return
	}
	w.RoomsLock.Lock()
	if peer, ok := w.Rooms[uuid]; ok {
		w.RoomsLock.Unlock()
		RoomViewerConn(c, peer.Peers)
		return
	}

	w.RoomsLock.Unlock()
}

func RoomViewerConn(c *websocket.Conn, p *w.Peers) {
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
