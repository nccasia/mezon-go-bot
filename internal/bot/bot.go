package bot

import (
	"ncc/go-mezon-bot/config"
	"ncc/go-mezon-bot/internal/constants"
	"strings"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

type IBot interface {
	StartCheckin()
	RegisterCommand(cmd string, handler CommandHandler)
}

type Bot struct {
	commands    map[string]CommandHandler
	logger      *zap.Logger
	signaling   mezonsdk.IWSConnection
	callService mezonsdk.ICallService
	checkinChan chan string
}

func NewBot(cfg *config.AppConfig, logger *zap.Logger) (IBot, error) {

	// make ws signaling
	signaling, err := mezonsdk.NewWSConnection(&mezonsdk.Config{
		BasePath:     cfg.Domain,
		ApiKey:       cfg.BotCheckin.ApiKey,
		Timeout:      15,
		InsecureSkip: cfg.InsecureSkip,
		UseSSL:       cfg.UseSSL,
	}, "")
	if err != nil {
		logger.Error("[NewBot] ws signaling error", zap.Error(err))
		return nil, err
	}

	// make call service
	callService := mezonsdk.NewCallService(cfg.BotCheckin.BotId, signaling, webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{constants.ICE},
	})

	return &Bot{
		commands:    make(map[string]CommandHandler),
		signaling:   signaling,
		callService: callService,
		logger:      logger,
		checkinChan: make(chan string, 1000),
	}, nil
}

func (b *Bot) RegisterCommand(cmd string, handler CommandHandler) {
	b.commands[cmd] = handler
}

func (b *Bot) handleCommand(input string) {
	// TODO: ws channel message
	parts := strings.Fields(input)
	if len(parts) == 0 {
		return
	}

	command := parts[0]
	args := parts[1:]

	if handler, exists := b.commands[command]; exists {
		handler(command, args)
	}
	return
}
