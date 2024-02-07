package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/websocket/v2"
	"log"
	"whatsapp-ki-maya/subscription"
)

func Connect(service subscription.Service) fiber.Handler {
	return websocket.New(func(conn *websocket.Conn) {
		apiKEY := conn.Headers("API_KEY")
		if apiKEY != "API_KEY" {
			conn.WriteMessage(websocket.TextMessage, []byte("invalid api key"))
			return
		}

		sub := conn.Query("sub")
		log.Println("CLIENT CONNECTED AT:", sub)

		client := subscription.NewClient(sub, conn)
		service.Subscribe() <- client
		go client.Listen(service)
		client.Serve()
	})
}