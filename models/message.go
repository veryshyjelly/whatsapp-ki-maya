package models

type Message struct {
	ChatId   string  `json:"chat_id"`
	Sender   string  `json:"sender"`
	Text     *string `json:"text,omitempty"`
	Image    []byte  `json:"photo,omitempty"`
	Video    []byte  `json:"video,omitempty"`
	Audio    []byte  `json:"audio,omitempty"`
	Document []byte  `json:"document,omitempty"`
	Sticker  []byte  `json:"sticker,omitempty"`
	Caption  *string `json:"caption"`
}

func (m *Message) GetChatId() string {
	return m.ChatId
}