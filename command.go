package main

import (
	"errors"
	"fmt"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/internal/helper"
	"mezon-go-bot/pkg/clients"
	"path/filepath"
	"strconv"
	"sync"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/api"
	"go.uber.org/zap"
)

var (
	ncc8AudioName string
	players       map[string]mezonsdk.AudioPlayer
	mu            sync.Mutex
)

func init() {
	players = make(map[string]mezonsdk.AudioPlayer)
}

func Ncc8Handler(command string, args []string, message *api.ChannelMessage) error {
	// Load Config
	cfg := config.LoadConfig()

	if len(args) == 0 || args[0] == "" {
		content := fmt.Sprintf("{\"t\":\"```Supported commands:   \\nCommand: *ncc8 play {ID} \\nCommand: *ncc8 stop    \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":83}]}")
		bot.SendMessage(message, content)
		return nil
	}
	switch args[0] {
	case constants.NCC8_ARG_PLAY:
		if len(args) > 1 {
			episodeID, err := strconv.Atoi(args[1])
			if err != nil {
				content := fmt.Sprintf("{\"t\":\"```Command: *ncc8 play {ID}   \\nExample: *ncc8 play 100   \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":58}]}")
				bot.SendMessage(message, content)
				bot.Logger().Error("[ncc8] args[1] is not a valid number", zap.Error(err))
				return nil
			}

			// apiURL := fmt.Sprintf("http://172.16.100.114:3000/ncc8/episode/%d", episodeID)

			// resp, err := http.Get(apiURL)
			// if err != nil {
			// 	bot.Logger().Error("[ncc8] failed to fetch episode URL", zap.Error(err))
			// 	return err
			// }

			// defer resp.Body.Close()

			// if resp.StatusCode != http.StatusOK {
			// 	content := fmt.Sprintf("{\"t\":\"Episode %d has not been released.\"}", episodeID)
			// 	bot.SendMessage(message, content)
			// 	bot.Logger().Error("[ncc8] non-200 response from API", zap.Int("statusCode", resp.StatusCode))
			// 	return err
			// }

			// var response struct {
			// 	URL string `json:"url"`
			// }
			// if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			// 	bot.Logger().Error("[ncc8] failed to parse response", zap.Error(err))
			// 	return err
			// }

			// targetFile := fmt.Sprintf("ncc8_%d.ogg", episodeID)

			// fileName, err := helper.FindFileByName(constants.NCC8_AUDIO_DIR, targetFile)
			fileName, err := helper.FindFileByName(constants.NCC8_AUDIO_DIR, "checkin-success.ogg")
			if err != nil {
				content := fmt.Sprintf("{\"t\":\"Episode %d has not been released.\"}", episodeID)
				bot.SendMessage(message, content)
			} else {
				content := fmt.Sprintf("{\"t\":\"NCC8 is broadcast on stream\",\"hg\":[{\"channelid\":\"%s\",\"s\":21,\"e\":40}]}", cfg.Ncc8ChannelId)
				bot.SendMessage(message, content)
			}

			ncc8AudioName = fileName
			// ncc8AudioName = response.URL

			player, _ := players[cfg.Ncc8ChannelId]
			player.Cancel(cfg.Ncc8ChannelId)

			return nil
		} else {
			content := fmt.Sprintf("{\"t\":\"```Supported commands:   \\nCommand: *ncc8 play {ID} \\nCommand: *ncc8 stop    \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":83}]}")
			bot.SendMessage(message, content)
		}

	case constants.NCC8_ARG_STOP:
		ncc8AudioName = ""
		player, _ := players[cfg.Ncc8ChannelId]
		player.Cancel(cfg.Ncc8ChannelId)
		content := "{\"t\":\"NCC8 broadcast has been stopped.\"}"
		bot.SendMessage(message, content)
		return nil
	}

	return nil
}

func HandlerPlayNCC8Default() error {
	cfg := config.LoadConfig()

	mu.Lock()
	player, exists := players[cfg.Ncc8ChannelId]
	if !exists {
		var err error
		player, err = mezonsdk.NewAudioPlayer(cfg.ClanId, cfg.Ncc8ChannelId, cfg.BotId, cfg.BotName, cfg.Token)
		if err != nil {
			mu.Unlock()
			bot.Logger().Error("[ncc8] cannot create player", zap.Error(err))
			return err
		}
		players[cfg.Ncc8ChannelId] = player
	}
	mu.Unlock()

	audioFiles, err := helper.GetAudioFiles(constants.NCC8_AUDIO_DIR, constants.NCC8_PREFIX)
	if err != nil || len(audioFiles) == 0 {
		bot.Logger().Error("[ncc8] failed to get audio files", zap.Error(err))
		return err
	}

	for {
		for _, file := range audioFiles {
			filePath := filepath.Join(constants.NCC8_AUDIO_DIR, file)
			if ncc8AudioName != "" {
				fileNCC8Path := filepath.Join(constants.NCC8_AUDIO_DIR, ncc8AudioName)
				err := player.Play(fileNCC8Path)
				if err != nil {
					bot.Logger().Error("[ncc8] failed to play audio from URL", zap.String("url", ncc8AudioName), zap.Error(err))
					return err
				}
				ncc8AudioName = ""
			} else {
				err := player.Play(filePath)
				if err != nil {
					bot.Logger().Error("[ncc8] failed to play audio file", zap.String("file", file), zap.Error(err))
					return err
				}
			}
		}
	}

}

func HandlerPlayDefault(channelId, botId, dir, prefix string) error {
	// Load Config
	cfg := config.LoadConfig()

	go func(channelId string) {
		mu.Lock()
		player, exists := players[channelId]
		if !exists {
			// Create a player if it doesn't exist yet
			var err error
			player, err = mezonsdk.NewAudioPlayer(cfg.ClanId, channelId, botId, cfg.BotName, cfg.Token)
			if err != nil {
				mu.Unlock()
				bot.Logger().Error("[ncc8] cannot create player", zap.Error(err))
				return
			}
			players[channelId] = player
		}
		mu.Unlock()

		// Get list of audio files
		audioFiles, err := helper.GetAudioFiles(dir, prefix)
		if err != nil {
			bot.Logger().Error("[ncc8] failed to get audio files", zap.Error(err))
			return
		}

		// Check if there is no file
		if len(audioFiles) == 0 {
			bot.Logger().Error("[ncc8] no audio files found")
			return
		}

		// Continuous Play
		for {
			for _, file := range audioFiles {
				filePath := filepath.Join(dir, file)
				err = player.Play(filePath)
				if err != nil {
					bot.Logger().Error("[ncc8] failed to play audio file", zap.String("file", file), zap.Error(err))
					return
				}
			}
		}
	}(channelId)

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
