package main

import (
	"encoding/json"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/internal/helper"
	"mezon-go-bot/internal/rtc"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"github.com/nccasia/mezon-go-sdk/configs"
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/api"
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/rtapi"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

type MsgContent struct {
	Content string `json:"t"`
}
type IBot interface {
	Start()
	Stop()
	RegisterCmd(prefix string, cmdHandler CommandHandler)
	Logger() *zap.Logger
	Config() *config.AppConfig
	MezonClient() *mezonsdk.Client
	SendMessage(message *api.ChannelMessage, content string) error
}

type Bot struct {
	cfg      *config.AppConfig
	commands map[string]CommandHandler
	logger   *zap.Logger
	mzn      *mezonsdk.Client

	// checkin
	callService rtc.ICallService
}

// MezonClient implements IBot.
func (b *Bot) MezonClient() *mezonsdk.Client {
	return b.mzn
}

// Logger implements IBot.
func (b *Bot) Logger() *zap.Logger {
	return b.logger
}

// Config implements IBot.
func (b *Bot) Config() *config.AppConfig {
	return b.cfg
}

// RegisterCmd implements IBot.
func (b *Bot) RegisterCmd(prefix string, cmdHandler CommandHandler) {
	if b.commands == nil {
		return
	}

	b.commands[prefix] = cmdHandler
}

// Stop implements IBot.
func (b *Bot) Stop() {
	panic("unimplemented")
}

func NewBot(cfg *config.AppConfig, logger *zap.Logger) (IBot, error) {

	// make ws signaling
	mzClient, err := mezonsdk.NewClient(&configs.Config{
		BasePath:     cfg.MznDomain,
		ApiKey:       cfg.ApiKey,
		Timeout:      15,
		InsecureSkip: cfg.InsecureSkip,
		UseSSL:       cfg.UseSSL,
	})
	if err != nil {
		logger.Error("[NewBot] ws signaling dm error", zap.Error(err))
		return nil, err
	}

	return &Bot{
		cfg:      cfg,
		commands: make(map[string]CommandHandler),
		mzn:      mzClient,
		logger:   logger,
	}, nil
}

func (b *Bot) Start() {
	socket, err := b.mzn.CreateSocket()
	if err != nil {
		b.logger.Error("[NewBot] can not create socket", zap.Error(err))
		return
	}
	socket.SetOnChannelMessage(func(e *rtapi.Envelope) error {
		go func() {
			err := b.handleCommand(e.GetChannelMessage())
			if err != nil {
				b.logger.Error("Error handling command", zap.Error(err))
			}
		}()
		return nil
	})

	callService := rtc.NewCallService(b.cfg.BotId, socket, webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{constants.ICE_MEZON},
	})
	socket.SetOnWebrtcSignalingFwd(callService.OnWebsocketEvent)
	callService.SetOnImage(CheckinHandler, constants.NUM_IMAGE_SNAPSHOT)
	callService.SetAcceptCallFileAudio(constants.CHECKIN_ACCEPT_CALL_AUDIO_PATH)
	callService.SetExitCallFileAudio(constants.CHECKIN_EXIT_CALL_AUDIO_PATH)
	callService.SetCheckinSuccessFileAudio(constants.CHECKIN_CHECKIN_SUCCESS_AUDIO_PATH)
	callService.SetCheckinFailFileAudio(constants.CHECKIN_CHECKIN_FAIL_AUDIO_PATH)

	HandlerPlayDefault(b.cfg.AudioBookChannelId, "456789", constants.BOOK_DIR, constants.BOOK_PREFIX)
	HandlerPlayNCC8Default()
}

type CommandHandler func(command string, args []string, msg *api.ChannelMessage) error

func (b *Bot) handleCommand(msg *api.ChannelMessage) error {
	content := msg.GetContent()
	if len(content) == 0 || len(content) >= 64 || content == "{}" {
		return nil
	}

	var msgContent *MsgContent
	if err := json.Unmarshal([]byte(content), &msgContent); err != nil {
		return err
	}

	if msgContent.Content != "" {
		command, args := helper.ExtractMessage(msgContent.Content)

		if handler, exists := b.commands[command]; exists {
			return handler(command, args, msg)
		}
	}

	return nil
}

func (b *Bot) SendMessage(message *api.ChannelMessage, content string) error {
	messageRef := &api.MessageRef{
		MessageRefId:             message.MessageId,
		Content:                  message.Content,
		MessageSenderId:          message.SenderId,
		MessageSenderUsername:    message.Username,
		MesagesSenderAvatar:      message.Avatar,
		MessageSenderDisplayName: message.DisplayName,
	}

	err := b.MezonClient().Socket.SendMessage(&rtapi.Envelope{
		Message: &rtapi.Envelope_ChannelMessageSend{
			ChannelMessageSend: &rtapi.ChannelMessageSend{
				ClanId:           message.ClanId,
				ChannelId:        message.ChannelId,
				Mode:             2,
				Content:          content,
				Mentions:         []*api.MessageMention{},
				Attachments:      []*api.MessageAttachment{},
				References:       []*api.MessageRef{messageRef},
				AnonymousMessage: false,
				MentionEveryone:  false,
				Avatar:           "",
				IsPublic:         true,
				Code:             0,
			},
		},
	})
	if err != nil {
		return err
	}
	return nil
}
