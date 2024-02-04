package subscription

import (
	"context"
	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/binary/proto"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"
	gproto "google.golang.org/protobuf/proto"
	"log"
	"net/http"
	"strings"
	"whatsapp-ki-maya/models"
	"whatsapp-ki-maya/pkg"
)

// Server is the abstraction for whatsapp or telegram etc.
// this interface handles all the updates that comes from server
// and should handle all the updates that needs to be sent to the server
type Server interface {
	Update() chan models.Message
	Listen(service Service)
	Serve()
}

type server struct {
	updates chan models.Message
	conn    *whatsmeow.Client
}

func NewServer(conn *whatsmeow.Client) Server {
	return &server{
		updates: make(chan models.Message, 100),
		conn:    conn,
	}
}

func (s *server) Update() chan models.Message {
	return s.updates
}

func (s *server) Listen(service Service) {
	s.conn.AddEventHandler(func(evt interface{}) {
		switch m := evt.(type) {
		case *events.Message:
			go func(m *events.Message) {
				if m == nil {
					return
				}

				if m.Info.IsGroup {
					info, err := s.conn.GetGroupInfo(m.Info.Chat)
					if err != nil {
						log.Println(err)
					}
					pkg.PrintMessage(m, info)
				} else {
					pkg.PrintMessage(m, nil)
				}

				//log.Println("Message received with ID:", m.Info.ID)

				update := models.Message{ChatId: m.Info.Chat.String(), Sender: m.Info.PushName}
				mess := m.Message

				file, err := []byte{}, error(nil)

				// nothing just download the file and assign to respective field of update
				switch {
				case mess.Conversation != nil:
				case mess.GetExtendedTextMessage().GetText() != "":
					text := mess.GetConversation() + mess.GetExtendedTextMessage().GetText()
					update.Text = &text
				case mess.ImageMessage != nil:
					file, err = s.conn.Download(mess.ImageMessage)
					update.Caption = mess.ImageMessage.Caption
					update.Image = file
				case mess.VideoMessage != nil:
					file, err = s.conn.Download(mess.VideoMessage)
					update.Caption = mess.VideoMessage.Caption
					update.Video = file
				case mess.DocumentMessage != nil:
					file, err = s.conn.Download(mess.DocumentMessage)
					update.Caption = mess.DocumentMessage.Caption
					update.Document = file
				case mess.AudioMessage != nil:
					file, err = s.conn.Download(mess.AudioMessage)
					update.Audio = file
				case mess.StickerMessage != nil:
					file, err = s.conn.Download(mess.StickerMessage)
					update.Sticker = file
				}
				if err != nil {
					return
				}

				service.SendToClients() <- update
			}(m)
		}
	})
}

// Serve methods sends the message to the server
func (s *server) Serve() {
	for {
		mess := <-s.updates
		var us = strings.Split(mess.ChatId, "@")
		jid := types.JID{
			User:   us[0],
			Server: us[1],
		}

		msg := new(proto.Message)

		text := mess.Sender + ": "
		var resp whatsmeow.UploadResponse
		var err error

		contextInfo := &proto.ContextInfo{
			Participant:   gproto.String("0@s.whatsapp.net"),
			QuotedMessage: &proto.Message{Conversation: gproto.String(mess.Sender)},
		}

		var caption string

		if mess.Caption != nil {
			caption = *mess.Caption
		}

		// nothing just uploading the file to the server and assigning the response to respective field of message
		switch {
		case mess.Text != nil:
			text += *mess.Text
			msg.Conversation = &text
		case mess.Image != nil:
			resp, err = s.conn.Upload(context.Background(), mess.Image, whatsmeow.MediaImage)
			msg.ImageMessage = &proto.ImageMessage{
				Url:           &resp.URL,
				Mimetype:      gproto.String(http.DetectContentType(mess.Image)),
				Caption:       gproto.String(caption),
				FileSha256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				MediaKey:      resp.MediaKey,
				FileEncSha256: resp.FileEncSHA256,
				DirectPath:    &resp.DirectPath,
				ContextInfo:   contextInfo,
			}
		case mess.Video != nil:
			resp, err = s.conn.Upload(context.Background(), mess.Image, whatsmeow.MediaVideo)
			msg.VideoMessage = &proto.VideoMessage{
				Url:           &resp.URL,
				Mimetype:      gproto.String(http.DetectContentType(mess.Video)),
				Caption:       gproto.String(caption),
				FileSha256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				MediaKey:      resp.MediaKey,
				FileEncSha256: resp.FileEncSHA256,
				DirectPath:    &resp.DirectPath,
				ContextInfo:   contextInfo,
			}
		case mess.Audio != nil:
			resp, err = s.conn.Upload(context.Background(), mess.Image, whatsmeow.MediaAudio)
			msg.AudioMessage = &proto.AudioMessage{
				Url:           &resp.URL,
				Mimetype:      gproto.String(http.DetectContentType(mess.Audio)),
				FileSha256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				MediaKey:      resp.MediaKey,
				FileEncSha256: resp.FileEncSHA256,
				DirectPath:    &resp.DirectPath,
				ContextInfo:   contextInfo,
			}
		case mess.Document != nil:
			resp, err = s.conn.Upload(context.Background(), mess.Image, whatsmeow.MediaDocument)
			msg.DocumentMessage = &proto.DocumentMessage{
				Url:           &resp.URL,
				Mimetype:      gproto.String(http.DetectContentType(mess.Document)),
				Title:         gproto.String("Document"),
				Caption:       gproto.String(caption),
				FileSha256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				MediaKey:      resp.MediaKey,
				FileEncSha256: resp.FileEncSHA256,
				DirectPath:    &resp.DirectPath,
				ContextInfo:   contextInfo,
			}
		case mess.Sticker != nil:
			resp, err = s.conn.Upload(context.Background(), mess.Sticker, whatsmeow.MediaImage)
			msg.StickerMessage = &proto.StickerMessage{
				Url:           &resp.URL,
				Mimetype:      gproto.String(http.DetectContentType(mess.Sticker)),
				FileSha256:    resp.FileSHA256,
				FileLength:    &resp.FileLength,
				MediaKey:      resp.MediaKey,
				FileEncSha256: resp.FileEncSHA256,
				DirectPath:    &resp.DirectPath,
				ContextInfo:   contextInfo,
			}
		case mess.Caption != nil:
			text += *mess.Caption
			mess.Caption = &text
		}
		if err != nil {
			log.Println("Error while creating message: ", err)
			continue
		}

		rsp, err := s.conn.SendMessage(context.Background(), jid, msg)
		log.Println("Message sent with ID:", rsp.ID)
	}
}