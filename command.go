package main

import (
	"errors"
	"mezon-go-bot/internal/constants"
	radiostation "mezon-go-bot/internal/radio-station"
	"mezon-go-bot/internal/rtc"
	"mezon-go-bot/pkg/clients"

	"github.com/nccasia/mezon-go-sdk/configs"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

func Ncc8Handler(command string, args []string) error {
	if args[0] == constants.NCC8_ARG_PLAY {
		wsConn, err := radiostation.NewWSConnection(&configs.Config{
			BasePath:     bot.Config().StnDomain,
			Timeout:      15,
			InsecureSkip: bot.Config().InsecureSkip,
			UseSSL:       bot.Config().UseSSL,
		}, "clanId", "channelId", bot.Config().BotId, "NCC8")
		if err != nil {
			bot.Logger().Error("[ncc8] radiostation new ws signaling error", zap.Error(err))
			return err
		}

		rtcConn, err := rtc.NewStreamingRTCConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{constants.ICE_GOOGLE},
		}, wsConn, "clanId", "channelId", bot.Config().BotId, "NCC8")
		if err != nil {
			bot.Logger().Error("[ncc8] new streaming rtc connection error", zap.Error(err))
			return err
		}

		// TODO: get mp3 by args[1]
		// TODO: ffmpeg convert mp3 to ogg: ffmpeg -i test.mp3 -c:a libopus -page_duration 20000 test.ogg
		err = rtcConn.SendAudioTrack("audio/ncc8.ogg")
		if err != nil {
			bot.Logger().Error("[ncc8] send audio file error", zap.Error(err))
			return err
		}
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
