package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
	"time"
	"whatsapp-ki-maya/subscription"
)

func Connect(service subscription.Service) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		err := conn.SetReadDeadline(time.Now().Add(time.Second * 10))
		if err != nil {
			return
		}

		if _, p, err := conn.ReadMessage(); err != nil {
			_ = conn.WriteMessage(1, []byte(err.Error()))
			return
		} else {
			x := string(p)
			if x != "API_KEY" {
				_ = conn.WriteMessage(1, []byte("Invalid API Key"))
				_ = conn.Close()
				return
			}
		}
		err = conn.SetReadDeadline(time.Time{})
		if err != nil {
			return
		}

		sub := conn.Query("sub")
		log.Println("client connected to subscription:", sub)

		client := subscription.NewClient(sub, conn)
		go client.Listen(service)
		client.Serve()
	})
}