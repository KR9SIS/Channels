package server

import (
	"flag"
	"os"
	"time"

	"channels/internal/handlers"
	w "channels/pkg/webrtc"
	"github.com/gofiber/contrib/v3/websocket"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/logger"
	"github.com/gofiber/template/html/v3"
)

var (
	addr = flag.String("addr", ":"+os.Getenv("PORT"), "")
	cert = flag.String("cert", "", "")
	key  = flag.String("key", "", "")
)

func Run() error {
	flag.Parse()

	if *addr == ":" {
		*addr = ":8080"
	}

	engine := html.New("./views", ".html")

	app := fiber.New(fiber.Config{Views: engine})

	app.Use(logger.New())
	app.Use(cors.New())

	app.Get("/", handlers.Welcome)

	app.Get("/room/create", handlers.RoomCreate)
	app.Get("/room/:uuid", handlers.Room)
	app.Get("/room/:uuid/websocket", websocket.New(
		handlers.RoomWebsocket, websocket.Config{HandshakeTimeout: 10 * time.Second},
	))
	app.Get("/room/:uuid/chat", handlers.RoomChat)
	app.Get("/room/:uuid/chat/websocket", websocket.New(handlers.RoomChatWebsocket))
	app.Get("/room/:uuid/viewer/websocket", websocket.New(handlers.RoomViewerWebsocket))

	app.Get("/stream/:suuid", handlers.Stream)
	app.Get("/stream/suuid/websocket", websocket.New(
		handlers.StreamWebSocket, websocket.Config{HandshakeTimeout: 10 * time.Second},
	))
	app.Get("/stream/:suuid/chat/websocket", websocket.New(handlers.StreamChatWebSocket))
	app.Get("/stream/:suuid/viewer/websocket", websocket.New(handlers.StreamViewerWebsocket))

	app.Static("/", "./assets")

	w.Rooms = make(map[string]*w.Room)
	w.Streams = make(map[string]*w.Room)

	go func() {
		for range time.NewTicker(time.Second * 3).C {
			for _, room := range w.Rooms {
				room.Peers.DispatchKeyFrame()
			}
		}
	}()

	if *cert != "" {
		return app.Server().ListenAndServeTLS(*addr, *cert, *key)
	}
	return app.Listen(*addr)
}
