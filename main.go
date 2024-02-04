package main

import (
	"fmt"
	"github.com/gofiber/fiber/v2"
	"log"
	"time"
	"whatsapp-ki-maya/api"
	"whatsapp-ki-maya/pkg"
	"whatsapp-ki-maya/subscription"
)

func main() {
	app := fiber.New()
	app.Get("/login", api.Login())
	go app.Listen(":8050")

	fmt.Println("Connecting to whatsapp")
	client := pkg.Connect("INFO")
	err := app.ShutdownWithTimeout(time.Second)
	if err != nil {
		log.Println("error shutting down previous server:", err)
	}
	server := subscription.NewServer(client)
	sub := subscription.NewService()
	sub.SetServer(server)
	sub.Run()

	app = fiber.New()
	app.Get("/connect", api.Connect(sub))
	log.Fatalln(app.Listen(":8050"))
}