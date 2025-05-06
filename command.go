package main

import (
	"errors"
	"fmt"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/pkg/clients"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"go.uber.org/zap"
)

//var audioPlayer mezonsdk.AudioPlayer

func Ncc8Handler(command string, args []string) error {
	cfg := config.LoadConfig()
	client, err := mezonsdk.NewClient(cfg.ApiKey)
	if err != nil {
		fmt.Println("error", err)
		return err
	}

	bot.Logger().Info("[ncc8] starts")
	audioPlayer, err := client.NewAudioPlayer(cfg.ClanId, cfg.ChannelId)
	if err != nil {
		bot.Logger().Error("[ncc8] create audio player error", zap.Error(err))
		return err
	}

	switch args[0] {
	case constants.NCC8_ARG_PLAY:

		// TODO: get mp3 by args[1]
		// TODO: ffmpeg convert mp3 to ogg: ffmpeg -i test.mp3 -c:a libopus -page_duration 20000 test.ogg
		err := audioPlayer.Play("https://raw.githubusercontent.com/mezonai/mezon-go-bot/refs/heads/main/audio/ncc8.ogg")
		if err != nil {
			bot.Logger().Error("[ncc8] play audio error", zap.Error(err))
			return err
		}

	case constants.NCC8_ARG_STOP:
		audioPlayer.Close(cfg.ChannelId)

	default:
		return errors.New("unknown command")
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
