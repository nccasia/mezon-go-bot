package main

import (
	"encoding/json"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/internal/helper"
	"mezon-go-bot/internal/rtc"
	"mezon-go-bot/internal/websocket"

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
	MezonClient() *mezonsdk.Client
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

	socket, err := mzClient.CreateSocket()
	if err != nil {
		logger.Error("[NewBot] can not create socket", zap.Error(err))
		return nil, err
	}

	callService := rtc.NewCallService(cfg.BotId, socket, webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{constants.ICE_MEZON},
	})
	socket.SetOnWebrtcSignalingFwd(callService.OnWebsocketEvent)
	callService.SetOnImage(CheckinHandler, constants.NUM_IMAGE_SNAPSHOT)
	callService.SetAcceptCallFileAudio(constants.CHECKIN_ACCEPT_CALL_AUDIO_PATH)
	callService.SetExitCallFileAudio(constants.CHECKIN_EXIT_CALL_AUDIO_PATH)
	callService.SetCheckinSuccessFileAudio(constants.CHECKIN_CHECKIN_SUCCESS_AUDIO_PATH)
	callService.SetCheckinFailFileAudio(constants.CHECKIN_CHECKIN_FAIL_AUDIO_PATH)

	return &Bot{
		cfg:      cfg,
		commands: make(map[string]CommandHandler),
		mzn:      mzClient,
		logger:   logger,
	}, nil
}

func (b *Bot) Start() {
	socket := b.mzn.Socket
	socket.SetOnChannelMessage(func(e *rtapi.Envelope) error {
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
