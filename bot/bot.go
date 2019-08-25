package bot

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/tevino/log"

	"github.com/awesome-memobird/the-memobird-bot/memobird"
	"github.com/awesome-memobird/the-memobird-bot/model"
	"github.com/awesome-memobird/the-memobird-bot/service"
	tb "gopkg.in/tucnak/telebot.v2"
)

// // Service represents the ability of the telegram bot service.
// type Service interface {
// 	BindDevice(telegramID int, deviceID string) error
// }

// UserService represents the ability of the user service.
type UserService interface {
	IsExistsByTelegramID(telegramID int) (bool, error)
	GetByTelegramID(telegramID int) (*model.User, error)
	New(*model.User) error
}

// DeviceService represents the ability of the device service.
type DeviceService interface {
	IsFree(memobirdID string) (bool, error)
	New(*model.Device) (*model.Device, error)
	GetByUserID(uint) (*model.Device, error)
	VerifyCodeByUserID(code string, userID uint) (bool, error)
}

// BirdService represents the ability of the bird service.
type BirdService interface {
	PrintTextToBird(birdID, text string) (*memobird.PrintResult, error)
	BindBirdWithMessage(birdID, msg string) (*memobird.PrintResult, error)
}

// Bot is a telegram bot.
type Bot struct {
	*Config
	*tb.Bot
}

// New creates a new telegram bot.
func New(config *Config) (*Bot, error) {
	rawBot, err := tb.NewBot(tb.Settings{
		Token:  config.Token,
		Poller: &tb.LongPoller{Timeout: config.PollerTimeout},
	})

	if err != nil {
		return nil, fmt.Errorf("error creating bot: %w", err)
	}
	b := &Bot{
		Bot:    rawBot,
		Config: config,
	}

	b.Handle(tb.OnText, b.handleText)
	return b, nil
}

const (
	replyFailedGettingData         = "I'm having trouble getting your data, please try again in a moment."
	replyMetBeforeS                = "Hi %s, we've met before."
	replyNiceToMeetYouS            = "Hello %s, nice to meet you!"
	replyCheckMemobirdID           = "Please check the Memobird ID provided."
	replyBindHelp                  = "Please use /bind [YourMemobirdID] to bind a Memobird before sending anything for printing"
	replyVerificationInstructionDS = `To complete the verification
please send:
    /verify %d
to:
    @%s
`
	replyFailedSendingVerification = "I'm having trouble sending you a verification code, please check the Memobird ID provided or try again in a moment."
	replyVerificationSent          = "A verification code with instructions was sent to your device, please follow it to complete the binding."
	replyBindComplete              = "Device binding complete!"
	replyVerificationFailed        = "Verification failed, please check the code or try again in a moment."
	replyFailedSendingMessageS     = "Error sending your message: %s"
	replySentPrintedTT             = "- Sent: %t\n- Printed: %t"
	replySent                      = "Sent"
	replySentFailure               = "The message failed to deliver"
)

func (b *Bot) handleStart(m *message) {
	b.Send(m.Sender, fmt.Sprintf(replyNiceToMeetYouS, m.SenderUser.TelegramFullName))
}

func (b *Bot) handleBind(m *message) {
	memobirdID := strings.TrimSpace(m.Payload)
	if memobirdID == "" {
		b.Send(m.Sender, replyBindHelp)
		return
	}

	isFree, err := b.DeviceService.IsFree(memobirdID)
	if err != nil {
		log.Warnf("error querying MemobirdID[%s]: %s", memobirdID, err)
		return
	}
	if !isFree {
		b.Send(m.Sender, replyCheckMemobirdID)
		return
	}

	device, err := b.DeviceService.New(&model.Device{
		UserID:     m.SenderUser.ID,
		MemobirdID: memobirdID,
	})
	if err != nil {
		log.Warnf("error creating device[%s] of user[%d]: %s", memobirdID, m.Sender.ID, err)
		return
	}
	_, err = b.BirdService.BindBirdWithMessage(memobirdID, fmt.Sprintf(replyVerificationInstructionDS,
		device.VerificationCode,
		b.Bot.Me.Username))
	if err != nil {
		log.Warnf("Error binding: %s", err)
		b.Send(m.Sender, replyFailedSendingVerification)
		return
	}
	b.Send(m.Sender, replyVerificationSent)
}

func (b *Bot) handleVerify(m *message) {
	verificationCode := strings.TrimSpace(m.Payload)
	if verificationCode == "" {
		b.Send(m.Sender, replyVerificationFailed)
		return
	}

	success, err := b.DeviceService.VerifyCodeByUserID(verificationCode, m.SenderUser.ID)
	if err != nil {
		log.Warnf("Error verifying user[%d] with code[%s]", m.SenderUser.ID, verificationCode)
		success = false
		return
	}
	var msg = replyVerificationFailed
	if success {
		msg = replyBindComplete
	}
	b.Send(m.Sender, msg)
}

func (b *Bot) handleSend(m *message) {
	reply := ""

	device, err := b.DeviceService.GetByUserID(m.SenderUser.ID)
	if err != nil && !service.IsRecordNotFoundError(err) {
		log.Warn("error querying device:", err)
		return
	}
	if service.IsRecordNotFoundError(err) {
		reply = replyBindHelp
	} else {
		result, err := b.BirdService.PrintTextToBird(device.MemobirdID, m.Payload)
		if err != nil {
			reply = fmt.Sprintf(replyFailedSendingMessageS, err)
		} else {
			if result.IsSuccess {
				reply = replySent
			} else {
				reply = replySentFailure
			}
		}
	}

	b.Send(m.Sender, reply, &tb.SendOptions{
		ReplyTo:   m.Message,
		ParseMode: tb.ModeMarkdown,
	})
}

func (b *Bot) handleText(msg *tb.Message) {
	if err := b.createUserIfNot(msg.Sender); err != nil {
		log.Warnf("Error creating telegram user[%s]: %s", msg.Sender.ID, err)
		b.Send(msg.Sender, replyFailedGettingData)
		return
	}
	m, err := b.wrapMessage(msg)
	if err != nil {
		log.Warnf("error wrapping message: %s", err)
		b.Send(m.Sender, replyFailedGettingData)
		return
	}

	switch m.Command {
	case "/start":
		b.handleStart(m)
	case "/verify":
		b.handleVerify(m)
	case "/bind":
		b.handleBind(m)
	case "/send":
		b.handleSend(m)
	default:
		b.handleSend(m)
	}
}

var reCmdPrefix = regexp.MustCompile(`^\/[a-z]+( .+)?`)

func (b *Bot) createUserIfNot(sender *tb.User) error {
	isUserExists, err := b.UserService.IsExistsByTelegramID(sender.ID)
	if err != nil {
		return fmt.Errorf("getting telegram user[%d]: %w", sender.ID, err)
	}

	if !isUserExists {
		fullName := strings.TrimSpace(sender.FirstName + " " + sender.LastName)
		b.UserService.New(&model.User{
			TelegramID:       int64(sender.ID),
			TelegramUserName: sender.Username,
			TelegramFullName: fullName,
		})
	} else {
		// TODO: update user profile
	}
	return nil
}

func splitCmdNPayload(txt string) (cmd, payload string) {
	if reCmdPrefix.MatchString(txt) {
		parts := strings.SplitN(txt, " ", 2)
		switch len(parts) {
		case 1:
			cmd = parts[0]
			payload = ""
		case 2:
			cmd = parts[0]
			payload = parts[1]
		default:
		}
	} else {
		payload = txt
	}
	return cmd, payload
}

func (b *Bot) wrapMessage(m *tb.Message) (*message, error) {
	user, err := b.UserService.GetByTelegramID(m.Sender.ID)
	if err != nil {
		return nil, fmt.Errorf("getting user by telegram ID[%d]: %w", m.Sender.ID, err)
	}
	cmd, payload := splitCmdNPayload(m.Text)
	return &message{
		Message:    m,
		SenderUser: user,
		Payload:    payload,
		Command:    cmd,
	}, nil
}
