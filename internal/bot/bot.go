package bot

import (
	"ncc/go-mezon-bot/config"
	radiostation "ncc/go-mezon-bot/internal/radio-station"
	"ncc/go-mezon-bot/internal/rtc"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"github.com/nccasia/mezon-go-sdk/configs"
	"go.uber.org/zap"
)

type IBot interface {
	Start()
	Setup() IBot

	NCC8() IBot
	Checkin() IBot
}

type Bot struct {
	cfg           *config.AppConfig
	botCfg        config.BotConfig
	commands      map[string]CommandHandler
	logger        *zap.Logger
	signalingDM   mezonsdk.IWSConnection
	signalingClan mezonsdk.IWSConnection

	// checkin
	callService rtc.ICallService

	// ncc8
	radioStationSignaling radiostation.IWSConnection
	streamingService      rtc.IStreamingRTCConnection
}

func NewBot(cfg *config.AppConfig, botCfg config.BotConfig, logger *zap.Logger) (IBot, error) {

	// make ws signaling
	signalingDM, err := mezonsdk.NewWSConnection(&configs.Config{
		BasePath:     cfg.Domain,
		ApiKey:       botCfg.ApiKey,
		Timeout:      15,
		InsecureSkip: cfg.InsecureSkip,
		UseSSL:       cfg.UseSSL,
	}, "")
	if err != nil {
		logger.Error("[NewBot] ws signaling dm error", zap.Error(err))
		return nil, err
	}

	signalingClan, err := mezonsdk.NewWSConnection(&configs.Config{
		BasePath:     cfg.Domain,
		ApiKey:       botCfg.ApiKey,
		Timeout:      15,
		InsecureSkip: cfg.InsecureSkip,
		UseSSL:       cfg.UseSSL,
	}, cfg.ClanId)
	if err != nil {
		logger.Error("[NewBot] ws signaling clan error", zap.Error(err))
		return nil, err
	}

	return &Bot{
		cfg:           cfg,
		botCfg:        botCfg,
		commands:      make(map[string]CommandHandler),
		signalingDM:   signalingDM,
		signalingClan: signalingClan,
		logger:        logger,
	}, nil
}

func (b *Bot) Start() {
	b.logger.Info("Bot is running...")
}
