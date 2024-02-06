package models

type Message struct {
	ChatId      string  `json:"chat_id"`
	Sender      string  `json:"sender"`
	Participant string  `json:"participant,omitempty"`
	Text        *string `json:"text,omitempty"`
	Image       []byte  `json:"photo,omitempty"`
	Video       []byte  `json:"video,omitempty"`
	Audio       []byte  `json:"audio,omitempty"`
	Document    []byte  `json:"document,omitempty"`
	Sticker     []byte  `json:"sticker,omitempty"`
	Filename    *string `json:"filename,omitempty"`
	Caption     *string `json:"caption,omitempty"`
}

func (m *Message) GetChatId() string {
	return m.ChatId
}