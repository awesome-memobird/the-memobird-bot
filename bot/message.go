package bot

import (
	"github.com/awesome-memobird/the-memobird-bot/model"
	tb "gopkg.in/tucnak/telebot.v2"
)

type message struct {
	*tb.Message
	SenderUser *model.User
	Payload    string
	Command    string
}

type ctxHandler func(m *message)
