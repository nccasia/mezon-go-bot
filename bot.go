package main

import (
	"encoding/json"
	"ncc/go-mezon-bot/config"
	"ncc/go-mezon-bot/internal/constants"
	"ncc/go-mezon-bot/internal/helper"
	"ncc/go-mezon-bot/internal/rtc"
	"ncc/go-mezon-bot/internal/websocket"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"github.com/nccasia/mezon-go-sdk/configs"
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/api"
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/rtapi"
	"github.com/pion/webrtc/v4"
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
	mzClient, err := mezonsdk.NewWSConnection(&configs.Config{
		BasePath:     cfg.MznDomain,
		ApiKey:       cfg.ApiKey,
		Timeout:      15,
		InsecureSkip: cfg.InsecureSkip,
		UseSSL:       cfg.UseSSL,
	}, "")
	if err != nil {
		logger.Error("[NewBot] ws signaling dm error", zap.Error(err))
		return nil, err
	}

	callService := rtc.NewCallService(cfg.BotId, mzClient, webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{constants.ICE_MEZON},
	})
	callService.SetOnImage(CheckinHandler, constants.NUM_IMAGE_SNAPSHOT)
	callService.SetAcceptCallFileAudio(constants.CHECKIN_ACCEPT_CALL_AUDIO_PATH)
	callService.SetExitCallFileAudio(constants.CHECKIN_EXIT_CALL_AUDIO_PATH)
	callService.SetCheckinSuccessFileAudio(constants.CHECKIN_CHECKIN_SUCCESS_AUDIO_PATH)
	callService.SetCheckinFailFileAudio(constants.CHECKIN_CHECKIN_FAIL_AUDIO_PATH)
	mzClient.SetOnWebrtcSignalingFwd(callService.OnWebsocketEvent)

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
