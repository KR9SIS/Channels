package handlers

import (
	"channels/pkg/chat"
	w "channels/pkg/webrtc"

	"github.com/fasthttp/websocket"
	"github.com/gofiber/fiber/v3"
)

func RoomChat(ctx fiber.Ctx) error {
	return ctx.Render("chat", fiber.Map{}, "layouts/main")
}

func RoomChatWebsocket(c *websocket.Conn) {
	uuid := c.Params("uuid")
	if uuid == "" {
		return
	}

	w.RoomsLock.Lock()
	room := w.Rooms[uuid]
	w.RoomsLock.Unlock()
	if room == nil || room.Hub == nil {
		return
	}
	chat.PeerChatConn(c, room.Hub)
}

func StreamChatWebsocket(c *websocket.Conn) {
	suuid := c.Params("suuid")
	if suuid == "" {
		return
	}
	w.RoomsLock.Lock()
	if stream, ok := w.Streams[suuid]; ok {
		w.RoomsLock.Unlock()
		if stream.Hub == nil {
			hub := chat.NewHub()
			stream.Hub = hub
			go hub.Run()
		}
		chat.PeerChatConn(c, stream.Hub)
		return
	}
	w.RoomsLock.Unlock()
}
