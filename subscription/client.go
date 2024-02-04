package subscription

import (
	"github.com/gofiber/websocket/v2"
	"log"
	"whatsapp-ki-maya/models"
)

type Client interface {
	Update() chan models.Message
	Subscription() string
	Listen(service Service)
	Serve()
}

type client struct {
	updates      chan models.Message
	subscription string
	Connection   *websocket.Conn
}

func NewClient(subscription string, conn *websocket.Conn) Client {
	return &client{
		subscription: subscription,
		Connection:   conn,
	}
}

func (c *client) Subscription() string {
	return c.subscription
}

func (c *client) Update() chan models.Message {
	return c.updates
}

func (c *client) Listen(service Service) {
	for {
		var update models.Message
		if err := c.Connection.ReadJSON(&update); err != nil {
			log.Println("error while reading message from client", err)
		}
		service.SendToServer() <- update
	}
}

func (c *client) Serve() {
	/*This function writes the updates to the client connection*/
	for {
		u := <-c.updates
		if err := c.Connection.WriteJSON(u); err != nil {
			log.Println("[ERROR] error occurred writing update to client", err)
			break
		}
	}
}