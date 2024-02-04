package pkg

import (
	"context"
	"fmt"
	"github.com/mattn/go-sqlite3"
	"github.com/skip2/go-qrcode"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store/sqlstore"
	waLog "go.mau.fi/whatsmeow/util/log"
	"os"
)

func Connect(minLevel string) *whatsmeow.Client {
	_, _, _ = sqlite3.Version()
	dbLog := waLog.Stdout("Database", minLevel, true)

	container, err := sqlstore.New("sqlite3", "file:whatsapp.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", minLevel, true)

	client := whatsmeow.NewClient(deviceStore, clientLog)
	//client.AddEventHandler(controllers.EventHandler)

	if client.Store.ID == nil { // No ID stored, new login
		qrChan, _ := client.GetQRChannel(context.Background())
		err = client.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				err = qrcode.WriteFile(evt.Code, qrcode.Medium, 256, "scan.png")
				fmt.Println("QR code:", evt.Code)
			} else {
				fmt.Println("Login event:", evt.Event)
				err = os.Remove("scan.png")
			}
		}
		if err != nil {
			panic(err)
		}
	} else { // Already logged in, just connect
		err = client.Connect()
		if err != nil {
			panic(err)
		}
	}

	return client
}