package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	"mezon-go-bot/internal/helper"
	"mezon-go-bot/pkg/clients"
	"net/http"
	"path/filepath"
	"sort"
	"strconv"
	"sync"
	"time"

	"github.com/robfig/cron/v3"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/api"
	"go.uber.org/zap"
)

var (
	ncc8AudioName      string
	players            map[string]mezonsdk.AudioPlayer
	mu                 sync.Mutex
	currentEpisodeNcc8 int
	isSchedule         bool
	isClose            bool
)

func init() {
	players = make(map[string]mezonsdk.AudioPlayer)
}

func Ncc8Handler(command string, args []string, message *api.ChannelMessage) error {
	// Load Config
	cfg := config.LoadConfig()

	if len(args) == 0 || args[0] == "" {
		content := fmt.Sprintf("{\"t\":\"```Supported commands:   \\nCommand: *ncc8 play {ID} \\nCommand: *ncc8 stop    \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":83}]}")
		bot.SendMessage(message, content, cfg.Ncc8ChannelId)
		return nil
	}
	switch args[0] {
	case constants.NCC8_ARG_PLAY:
		if len(args) > 1 {
			episodeID, err := strconv.Atoi(args[1])
			if err != nil {
				content := fmt.Sprintf("{\"t\":\"```Command: *ncc8 play {ID}   \\nExample: *ncc8 play 100   \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":58}]}")
				bot.SendMessage(message, content, cfg.Ncc8ChannelId)
				bot.Logger().Error("[ncc8] args[1] is not a valid number", zap.Error(err))
				return nil
			}

			if !isSchedule && currentEpisodeNcc8 == episodeID {
				episodeText := fmt.Sprintf("NCC8 số %d đang được phát trên stream", episodeID)
				length := len(episodeText)
				content := fmt.Sprintf("{\"t\":\"%s\",\"hg\":[{\"channelid\":\"%s\",\"s\":%d,\"e\":%d}]}", episodeText, cfg.Ncc8ChannelId, length-6, length+10)
				bot.SendMessage(message, content, cfg.Ncc8ChannelId)
				return nil
			}

			apiURL := fmt.Sprintf("http://172.16.100.114:3000/ncc8/episode/%d", episodeID)

			resp, err := http.Get(apiURL)
			if err != nil {
				content := fmt.Sprintf("{\"t\":\"NCC8 số %d không có trong danh sách.\"}", episodeID)
				bot.SendMessage(message, content, cfg.Ncc8ChannelId)
				bot.Logger().Error("[ncc8] failed to fetch episode URL", zap.Error(err))
				return nil
			}

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				content := fmt.Sprintf("{\"t\":\"NCC8 số %d không có trong danh sách.\"}", episodeID)
				bot.SendMessage(message, content, cfg.Ncc8ChannelId)
				bot.Logger().Error("[ncc8] non-200 response from API", zap.Int("statusCode", resp.StatusCode))
				return nil
			}

			var response struct {
				URL string `json:"url"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				content := fmt.Sprintf("{\"t\":\"NCC8 số %d không có trong danh sách.\"}", episodeID)
				bot.SendMessage(message, content, cfg.Ncc8ChannelId)
				bot.Logger().Error("[ncc8] failed to parse response", zap.Error(err))
				return nil
			}

			targetFile := fmt.Sprintf("ncc8_%d.ogg", episodeID)

			fileName, err := helper.FindFileByName(cfg.Ncc8AudioDir, targetFile)
			// fileName, err := helper.FindFileByName(constants.NCC8_AUDIO_DIR, "ncc8_206.ogg")
			if err != nil {
				content := fmt.Sprintf("{\"t\":\"NCC8 số %d không có trong danh sách.\"}", episodeID)
				bot.SendMessage(message, content, cfg.Ncc8ChannelId)
				return nil
			} else {
				currentEpisodeNcc8 = episodeID
				ncc8AudioName = fileName
				// ncc8AudioName = response.URL
				player, _ := players[cfg.Ncc8ChannelId]
				player.Cancel(cfg.Ncc8ChannelId)
				fileNCC8Path := filepath.Join(cfg.Ncc8AudioDir, ncc8AudioName)
				episodeText := fmt.Sprintf("NCC8 số %d đang được phát trên stream", episodeID)
				length := len(episodeText)
				content := fmt.Sprintf("{\"t\":\"%s\",\"hg\":[{\"channelid\":\"%s\",\"s\":%d,\"e\":%d}]}", episodeText, cfg.Ncc8ChannelId, length-6, length+10)
				bot.SendMessage(message, content, cfg.Ncc8ChannelId)
				err = player.Play(fileNCC8Path)
				if err != nil {
					bot.Logger().Error("[ncc8] failed to play audio from URL", zap.String("url", ncc8AudioName), zap.Error(err))
					return err
				}
				content = "{\"t\":\"NCC8 phát sóng theo số đã kết thúc. Phát sóng tự động sẽ được phát sau 3s.\"}"
				bot.SendMessage(message, content, cfg.Ncc8ChannelId)
				time.Sleep(3 * time.Second)
				currentEpisodeNcc8 = 0
				ncc8AudioName = ""
				player.Cancel(cfg.Ncc8ChannelId)
				return nil
			}
		} else {
			content := fmt.Sprintf("{\"t\":\"```Supported commands:   \\nCommand: *ncc8 play {ID} \\nCommand: *ncc8 stop    \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":83}]}")
			bot.SendMessage(message, content, cfg.Ncc8ChannelId)
			return nil
		}

	case constants.NCC8_ARG_STOP:
		if currentEpisodeNcc8 > 0 {
			ncc8AudioName = ""
			player, _ := players[cfg.Ncc8ChannelId]
			player.Cancel(cfg.Ncc8ChannelId)
			content := "{\"t\":\"NCC8 theo số đã ngừng phát sóng.\"}"
			bot.SendMessage(message, content, cfg.Ncc8ChannelId)
			return nil
		} else {
			content := "{\"t\":\"NCC8 theo số chưa được phát sóng.\"}"
			bot.SendMessage(message, content, cfg.Ncc8ChannelId)
			return nil
		}

	case constants.NCC8_ARG_PLAYLIST:
		apiURL := "http://172.16.100.114:3000/getAllNcc8Playlist"

		resp, err := http.Get(apiURL)
		if err != nil {
			bot.Logger().Error("[ncc8] failed to fetch playlist", zap.Error(err))
			return err
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			content := "{\"t\":\"Failed to fetch NCC8 playlist. Please try again later.\"}"
			bot.SendMessage(message, content, cfg.Ncc8ChannelId)
			return fmt.Errorf("non-200 response from API: %d", resp.StatusCode)
		}

		// Parse JSON
		var response []struct {
			Episode int `json:"episode"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
			bot.Logger().Error("[ncc8] failed to parse playlist JSON", zap.Error(err))
			return err
		}

		// Get episode list and remove duplicates
		var episodes []int
		for _, item := range response {
			episodes = append(episodes, item.Episode)
		}
		episodes = helper.RemoveDuplicatesFile(episodes)

		// Sort the episode list
		sort.Ints(episodes)

		// Divide the list into groups
		maxBatchSize := 100
		episodeBatches := helper.SplitEpisodesByRoundNumber(episodes, maxBatchSize)

		for _, batch := range episodeBatches {
			var content string
			for i, ep := range batch {
				if i > 0 {
					content += ", "
				}
				content += strconv.Itoa(ep)
			}

			formattedContent := fmt.Sprintf("{\"t\":\"```Danh sách NCC8:\\n%s\",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":%d}]}", content, len(content)+20)
			bot.SendMessage(message, formattedContent, cfg.Ncc8ChannelId)
		}

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

	audioFiles, err := helper.GetAudioFiles(cfg.Ncc8AudioDir, cfg.Ncc8Prefix)
	if err != nil || len(audioFiles) == 0 {
		bot.Logger().Error("[ncc8] failed to get audio files", zap.Error(err))
		return err
	}

	for {
		for _, file := range audioFiles {
			filePath := filepath.Join(cfg.Ncc8AudioDir, file)
			if ncc8AudioName == "" && isClose == false {
				// content := fmt.Sprintf("{\"t\":\"Đang phát: %s\"}", file)
				// bot.SendMessage(nil, content, cfg.Ncc8ChannelId)
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
				if isClose == false {
					// content := fmt.Sprintf("{\"t\":\"Đang phát: %s\"}", file)
					// bot.SendMessage(nil, content, cfg.AudioBookChannelId)
					filePath := filepath.Join(dir, file)
					err = player.Play(filePath)
					if err != nil {
						bot.Logger().Error("[ncc8] failed to play audio file", zap.String("file", file), zap.Error(err))
						return
					}
				}
			}
		}
	}(channelId)

	return nil
}

func HandleClosePlayer() {
	cfg := config.LoadConfig()
	isClose = true
	mu.Lock()
	defer mu.Unlock()

	// Close player
	for _, channelId := range []string{cfg.Ncc8ChannelId} {
		if player, exists := players[channelId]; exists {
			fmt.Printf("Closing player for channel %s...\n", channelId)
			player.Cancel(channelId)
		} else {
			fmt.Printf("No player found for channel %s.\n", channelId)
		}
	}
}

func ScheduleFridayAudio() {
	cfg := config.LoadConfig()

	// Create the scheduler
	c := cron.New()
	mu.Lock()
	player, exists := players[cfg.Ncc8ChannelId]
	if !exists {
		var err error
		player, err = mezonsdk.NewAudioPlayer(cfg.ClanId, cfg.Ncc8ChannelId, cfg.BotId, cfg.BotName, cfg.Token)
		if err != nil {
			mu.Unlock()
			bot.Logger().Error("[ncc8] cannot create player", zap.Error(err))
		}
		players[cfg.Ncc8ChannelId] = player
	}
	mu.Unlock()
	// Schedule the task to run every Friday at 11:30 AM +7 (04:30 AM UTC)
	_, err := c.AddFunc("30 4 * * 5", func() {
		// isSchedule = true
		// ncc8AudioName = "1111"

		// player.Cancel(cfg.Ncc8ChannelId)

		// Prepare audio file and episode details
		// fileNCC8Path := filepath.Join(constants.NCC8_AUDIO_DIR, ncc8AudioName)
		episodeText := fmt.Sprintf("NCC8 số %d đang được phát trên ", 219)
		length := len(episodeText)
		content := fmt.Sprintf("{\"t\":\"%s\",\"hg\":[{\"channelid\":\"%s\",\"s\":%d,\"e\":%d}]}", episodeText, cfg.Ncc8ChannelId, length, length+10)

		// Send initial message
		bot.SendMessage(nil, content, cfg.Ncc8ChannelId)

		// Play the audio
		err := player.Play("./audio/ncc8_219.ogg")
		if err != nil {
			bot.Logger().Error("[ncc8] failed to play audio from URL", zap.String("url", ncc8AudioName), zap.Error(err))
			return
		}

		// Send final message after 3 seconds
		// content = "{\"t\":\"NCC8 phát sóng theo số đã kết thúc. Phát sóng tự động sẽ được phát sau 3s.\"}"
		content = "{\"t\":\"NCC8 phát sóng hàng tuần đã kết thúc.\"}"

		bot.SendMessage(nil, content, cfg.Ncc8ChannelId)

		// Graceful sleep before resetting values
		// time.Sleep(3 * time.Second)

		// Reset variables and cancel the player
		// isSchedule = false
		// ncc8AudioName = ""
		player.Cancel(cfg.Ncc8ChannelId)
	})

	if err != nil {
		bot.Logger().Error("[ncc8] failed to schedule cron job", zap.Error(err))
	}

	// Start the scheduler
	c.Start()
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
