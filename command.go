package main

import (
	"errors"
	"fmt"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/pkg/clients"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/api"
	"go.uber.org/zap"
)

func Ncc8Handler(command string, args []string, message *api.ChannelMessage) error {
	cfg := config.LoadConfig()
	cfg.ClanId = "1775731152322039808"
	cfg.ChannelId = "1840654626240598016"
	client, err := mezonsdk.NewClient(cfg.ApiKey)
	if err != nil {
		fmt.Println("error", err)
		return err
	}

	bot.Logger().Info("[ncc8] starts")
	audioPlayer, err := client.NewAudioPlayer(cfg.ClanId, cfg.ChannelId)

	if len(args) == 0 || args[0] == "" {
		content := fmt.Sprintf("{\"t\":\"```Supported commands:   \\nCommand: *ncc8 play {ID} \\nCommand: *ncc8 stop    \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":83}]}")
		bot.SendMessage(message, content)
		return nil
	}

	switch args[0] {
	case constants.NCC8_ARG_PLAY:
		content := fmt.Sprintf("{\"t\":\"playing...\"}")
		err := audioPlayer.Play("https://raw.githubusercontent.com/mezonai/mezon-go-bot/refs/heads/main/audio/ncc8.ogg")
		if err != nil {
			bot.Logger().Error("[ncc8] play audio error", zap.Error(err))
			return err
		}
		bot.SendMessage(message, content)

	case constants.NCC8_ARG_STOP:
		content := "{\"t\":\"NCC8 has not been broadcast.\"}"
		bot.SendMessage(message, content)

	default:
		content := fmt.Sprintf("{\"t\":\"```Supported commands:   \\nCommand: *ncc8 play {ID} \\nCommand: *ncc8 stop    \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":83}]}")
		bot.SendMessage(message, content)
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
