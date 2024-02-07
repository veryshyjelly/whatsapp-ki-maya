package subscription

import (
	"context"
	"fmt"
	"github.com/emersion/go-vcard"
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

				if m.Message.GetConversation() == ".id" {
					m.Message.Conversation = gproto.String(m.Info.Chat.String())
					s.conn.SendMessage(context.Background(), m.Info.Chat, &proto.Message{Conversation: gproto.String(m.Info.Chat.String())})
					return
				}

				//log.Println("Message received with ID:", m.Info.ID)
				if !service.HasSubscribers(m.Info.Chat.String()) {
					log.Println("No subscribers for this chat: ", m.Info.Chat.String())
					return
				} else {
					//log.Println("Sending message to clients: ", m.Info.Chat.String())
				}

				update := models.Message{ChatId: m.Info.Chat.String(), Sender: m.Info.PushName, Participant: m.Info.Sender.String()}
				mess := m.Message

				file, err := []byte{}, error(nil)
				// nothing just download the file and assign to respective field of update
				switch {
				case mess.Conversation != nil:
					log.Println("conversation")
					update.Text = new(string)
					*update.Text = mess.GetConversation() + mess.GetExtendedTextMessage().GetText()
				case mess.GetExtendedTextMessage().GetText() != "":
					log.Println("extended text")
					update.Text = new(string)
					*update.Text = mess.GetConversation() + mess.GetExtendedTextMessage().GetText()
				case mess.ImageMessage != nil:
					log.Println("image message")
					file, err = s.conn.Download(mess.ImageMessage)
					update.Caption = mess.ImageMessage.Caption
					update.Image = file
				case mess.VideoMessage != nil:
					log.Println("video message")
					file, err = s.conn.Download(mess.VideoMessage)
					update.Caption = mess.VideoMessage.Caption
					update.Video = file
				case mess.DocumentMessage != nil:
					log.Println("document message")
					file, err = s.conn.Download(mess.DocumentMessage)
					update.Caption = mess.DocumentMessage.Caption
					update.Filename = mess.DocumentMessage.FileName
					update.Document = file
				case mess.AudioMessage != nil:
					log.Println("audio message")
					file, err = s.conn.Download(mess.AudioMessage)
					update.Audio = file
				case mess.StickerMessage != nil:
					log.Println("sticker message")
					file, err = s.conn.Download(mess.StickerMessage)
					update.Sticker = file
				case mess.ContactMessage != nil:
					log.Println("Contact message: ", mess.ContactMessage)
					if mess.ContactMessage.Vcard != nil {
						dec, _ := vcard.NewDecoder(strings.NewReader(*mess.ContactMessage.Vcard)).Decode()
						fmt.Println("Contact message: ", dec)
					}
				default:
					return
				}

				if mess.ExtendedTextMessage.ContextInfo.QuotedMessage != nil {
					update.QuotedText = gproto.String(mess.ExtendedTextMessage.ContextInfo.QuotedMessage.GetConversation())
				}

				if err != nil {
					fmt.Println("error while downloading file: ", err)
					return
				}

				//fmt.Println("Update created: ", update)
				service.SendToClients() <- update
			}(m)
		}
	})
}

// Serve methods sends the message to the server
func (s *server) Serve() {
	for mess := range s.updates {
		var us = strings.Split(mess.ChatId, "@")
		jid := types.JID{
			User:   us[0],
			Server: us[1],
		}

		msg := new(proto.Message)

		text := "*" + mess.Sender + "*: "
		var resp whatsmeow.UploadResponse
		var err error

		var participant *string
		if mess.Participant != "" {
			participant = gproto.String(mess.Participant)
		} else {
			participant = gproto.String("0@s.whatsapp.net")
		}

		contextInfo := &proto.ContextInfo{
			Participant:   participant,
			QuotedMessage: &proto.Message{Conversation: gproto.String(mess.Sender)},
		}
		if mess.QuotedText != nil {
			contextInfo.QuotedMessage = &proto.Message{Conversation: gproto.String(*mess.QuotedText)}
		} else if mess.Sticker == nil {
			contextInfo = nil
		}

		var caption string
		if mess.Caption != nil && strings.TrimSpace(*mess.Caption) != "" {
			caption = mess.Sender + ": " + *mess.Caption
		}

		// nothing just uploading the file to the server and assigning the response to respective field of message
		switch {
		case mess.Text != nil:
			text += *mess.Text
			if mess.QuotedText != nil {
				msg.ExtendedTextMessage = &proto.ExtendedTextMessage{
					Text:        gproto.String(text),
					ContextInfo: contextInfo,
				}
			} else {
				msg.Conversation = &text
			}

		case mess.Image != nil && len(mess.Image) > 0:
			resp, err = s.conn.Upload(context.Background(), mess.Image, whatsmeow.MediaImage)
			if caption == "" {
				caption = mess.Sender + " sent an image"
			}
			msg.ImageMessage = &proto.ImageMessage{
				Url:                 &resp.URL,
				Mimetype:            gproto.String(http.DetectContentType(mess.Image)),
				Caption:             gproto.String(caption),
				FileSha256:          resp.FileSHA256,
				FileLength:          &resp.FileLength,
				MediaKey:            resp.MediaKey,
				FileEncSha256:       resp.FileEncSHA256,
				DirectPath:          &resp.DirectPath,
				ContextInfo:         contextInfo,
				JpegThumbnail:       mess.Image,
				ThumbnailDirectPath: &resp.DirectPath,
				ThumbnailEncSha256:  resp.FileEncSHA256,
				ThumbnailSha256:     resp.FileEncSHA256,
			}
		case mess.Video != nil && len(mess.Video) > 0:
			resp, err = s.conn.Upload(context.Background(), mess.Image, whatsmeow.MediaVideo)
			if caption == "" {
				caption = mess.Sender + " sent a video"
			}
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
		case mess.Audio != nil && len(mess.Audio) > 0:
			resp, err = s.conn.Upload(context.Background(), mess.Audio, whatsmeow.MediaAudio)
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
		case mess.Document != nil && len(mess.Document) > 0:
			resp, err = s.conn.Upload(context.Background(), mess.Document, whatsmeow.MediaDocument)
			if caption == "" {
				caption = mess.Sender + " sent a document"
			}
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
		case mess.Sticker != nil && len(mess.Sticker) > 0:
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
			msg.Conversation = gproto.String(*mess.Caption)
		}
		if err != nil {
			log.Println("Error while creating message: ", err)
			continue
		}

		rsp, err := s.conn.SendMessage(context.Background(), jid, msg)
		if err != nil {
			log.Println("Error while sending message: ", err)
			continue
		}
		log.Println("Message sent with ID:", rsp.ID)
	}
}