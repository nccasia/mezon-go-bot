package main

import (
	"encoding/json"
	"ncc/go-mezon-bot/config"
	"ncc/go-mezon-bot/internal/helper"
	"ncc/go-mezon-bot/internal/rtc"
	"ncc/go-mezon-bot/internal/websocket"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"github.com/nccasia/mezon-go-sdk/configs"
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/api"
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/rtapi"
	"go.uber.org/zap"
)

type IBot interface {
	Start()
	Stop()

	RegisterCmd(prefix string, cmdHandler CommandHandler)

	Logger() *zap.Logger

	Config() *config.AppConfig
}

type Bot struct {
	cfg      *config.AppConfig
	commands map[string]CommandHandler
	logger   *zap.Logger
	mzClient mezonsdk.IWSConnection

	// checkin
	callService rtc.ICallService
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
	panic("unimplemented")
}

// Stop implements IBot.
func (b *Bot) Stop() {
	panic("unimplemented")
}

func NewBot(cfg *config.AppConfig, logger *zap.Logger) (IBot, error) {

	// make ws signaling
	mzClient, err := mezonsdk.NewWSConnection(&configs.Config{
		BasePath:     cfg.Domain,
		ApiKey:       cfg.ApiKey,
		Timeout:      15,
		InsecureSkip: cfg.InsecureSkip,
		UseSSL:       cfg.UseSSL,
	}, "")
	if err != nil {
		logger.Error("[NewBot] ws signaling dm error", zap.Error(err))
		return nil, err
	}

	return &Bot{
		cfg:      cfg,
		commands: make(map[string]CommandHandler),
		mzClient: mzClient,
		logger:   logger,
	}, nil
}

func (b *Bot) Start() {
	b.mzClient.SetOnChannelMessage(func(e *rtapi.Envelope) error {
		return b.handleCommand(e.GetChannelMessage())
	})
}

type CommandHandler func(command string, args []string) error

func (b *Bot) handleCommand(msg *api.ChannelMessage) error {

	var msgContent *websocket.MsgContent
	if err := json.Unmarshal([]byte(msg.GetContent()), &msgContent); err != nil || msgContent == nil {
		return err
	}

	command, args := helper.ExtractMessage(msgContent.Content)
	b.logger.Debug("[ExtractMessage]", zap.String("command", command), zap.Any("args", args))
	if handler, exists := b.commands[command]; exists {
		return handler(command, args)
	}
	return nil
}
