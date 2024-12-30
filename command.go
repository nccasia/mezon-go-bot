package main

import (
	"errors"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/internal/rtc"
	"mezon-go-bot/internal/ws"
	"mezon-go-bot/pkg/clients"

	"github.com/nccasia/mezon-go-sdk/configs"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

func Ncc8Handler(command string, args []string) error {
	// Load Config
	cfg := config.LoadConfig()

	wsConn, err := ws.NewWSConnection(&configs.Config{
		BasePath:     bot.Config().StnDomain,
		Timeout:      15,
		InsecureSkip: bot.Config().InsecureSkip,
		UseSSL:       false,
	}, cfg.ClanId, cfg.ChannelId, bot.Config().BotId, cfg.BotName, cfg.Token)
	if err != nil {
		bot.Logger().Error("[ncc8] radiostation new ws signaling error", zap.Error(err))
		return err
	}

	switch args[0] {
	case constants.NCC8_ARG_PLAY:
		rtcConn, err := rtc.NewStreamingRTCConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{constants.ICE_GOOGLE},
		}, wsConn, cfg.ClanId, cfg.ChannelId, bot.Config().BotId, "NCC8")
		if err != nil {
			bot.Logger().Error("[ncc8] new streaming rtc connection error", zap.Error(err))
			return err
		}

		rtcConn.Start()

	case constants.NCC8_ARG_STOP:
		//rtcConn.Stop()

	}

	return nil
}

func CheckinHandler(imageBase64 string) error {
	res, err := clients.CheckinApi(imageBase64)
	if err != nil {
		bot.Logger().Error("[CheckinApi] error", zap.Error(err))
		return err
	}

	bot.Logger().Info("[CheckinApi] send image", zap.Any("info", res))
	if res.Probability >= constants.CHECKIN_PROBABILITY_SUCCESS {
		bot.Logger().Info("[CheckinApi] checkin success", zap.Any("info", res))

		// return error close send image base64 to function
		return errors.New("checkin success")
	}

	// return nil -> continue send image base64 to function
	return nil
}
