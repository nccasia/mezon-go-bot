package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"mezon-go-bot/config"
	"mezon-go-bot/internal/constants"
	radiostation "mezon-go-bot/internal/radio-station"
	"mezon-go-bot/internal/rtc"
	"mezon-go-bot/pkg/clients"
	"net/http"
	"strconv"

	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/api"

	"github.com/nccasia/mezon-go-sdk/configs"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

var isPlaying bool

func Ncc8Handler(command string, args []string, message *api.ChannelMessage) error {
	// Load Config
	cfg := config.LoadConfig()

	wsConn, err := radiostation.NewWSConnection(&configs.Config{
		BasePath:     bot.Config().StnDomain,
		Timeout:      15,
		InsecureSkip: bot.Config().InsecureSkip,
		UseSSL:       false,
	}, cfg.ClanId, cfg.ChannelId, bot.Config().BotId, cfg.BotName, cfg.Token)
	if err != nil {
		bot.Logger().Error("[ncc8] radiostation new ws signaling error", zap.Error(err))
		return err
	}

	if len(args) == 0 || args[0] == "" {
		content := fmt.Sprintf("{\"t\":\"```Supported commands:   \\nCommand: *ncc8 play {ID} \\nCommand: *ncc8 stop    \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":83}]}")
		bot.SendMessage(message, content)
		return nil
	}

	switch args[0] {
	case constants.NCC8_ARG_PLAY:
		if len(args) > 1 {

			if isPlaying {
				content := fmt.Sprintf("{\"t\":\"NCC8 has been broadcast on stream\",\"hg\":[{\"channelid\":\"%s\",\"s\":27,\"e\":40}]}", cfg.ChannelId)
				bot.SendMessage(message, content)
				return nil
			}

			episodeID, err := strconv.Atoi(args[1])
			if err != nil {
				content := fmt.Sprintf("{\"t\":\"```Command: *ncc8 play {ID}   \\nExample: *ncc8 play 100   \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":58}]}")
				bot.SendMessage(message, content)
				bot.Logger().Error("[ncc8] args[1] is not a valid number", zap.Error(err))
				return nil
			}

			apiURL := fmt.Sprintf("http://172.16.100.114:3000/ncc8/episode/%d", episodeID)

			resp, err := http.Get(apiURL)
			if err != nil {
				bot.Logger().Error("[ncc8] failed to fetch episode URL", zap.Error(err))
				return err
			}

			bot.Logger().Info("resp", zap.Any("resp", resp))

			defer resp.Body.Close()

			if resp.StatusCode != http.StatusOK {
				content := fmt.Sprintf("{\"t\":\"Episode %d has not been released.\"}", episodeID)
				bot.SendMessage(message, content)
				bot.Logger().Error("[ncc8] non-200 response from API", zap.Int("statusCode", resp.StatusCode))
				return err
			}

			var response struct {
				URL string `json:"url"`
			}
			if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
				bot.Logger().Error("[ncc8] failed to parse response", zap.Error(err))
				return err
			}

			rtcConn, err := rtc.NewStreamingRTCConnection(webrtc.Configuration{
				ICEServers: []webrtc.ICEServer{constants.ICE_GOOGLE},
			}, wsConn, cfg.ClanId, cfg.ChannelId, bot.Config().BotId, "NCC8")
			if err != nil {
				bot.Logger().Error("[ncc8] new streaming RTC connection error", zap.Error(err))
				return err
			}

			isPlaying = true

			err = rtcConn.SendAudioTrack(response.URL)
			// err = rtcConn.SendAudioTrack("audio/lk_thucuoi.ogg")
			content := fmt.Sprintf("{\"t\":\"NCC8 is broadcast on stream\",\"hg\":[{\"channelid\":\"%s\",\"s\":21,\"e\":40}]}", cfg.ChannelId)
			bot.SendMessage(message, content)
			if err != nil {
				bot.Logger().Error("[ncc8] failed to send audio file", zap.Error(err))
				return err
			}

			return nil
		} else {
			content := fmt.Sprintf("{\"t\":\"```Supported commands:   \\nCommand: *ncc8 play {ID} \\nCommand: *ncc8 stop    \",\"mk\":[{\"type\":\"t\",\"s\":0,\"e\":83}]}")
			bot.SendMessage(message, content)
		}

	case constants.NCC8_ARG_STOP:
		isPlaying = false
		content := "{\"t\":\"NCC8 has not been broadcast.\"}"
		rtcConn, ok := rtc.MapStreamingRtcConn.Load(cfg.ChannelId)
		if !ok {
			bot.Logger().Error("Connection not found for channelId", zap.String("channelId", cfg.ChannelId))
		}

		if conn, ok := rtcConn.(*rtc.StreamingRTCConn); ok {
			content = "{\"t\":\"NCC8 has stopped broadcasting.\"}"
			isPlaying = false
			conn.Close(cfg.ChannelId)
			bot.Logger().Info("Connection closed for channelId", zap.String("channelId", cfg.ChannelId))
		} else {
			bot.Logger().Error("Error casting connection", zap.String("channelId", cfg.ChannelId))
		}

		bot.SendMessage(message, content)

		_, ok = rtc.MapStreamingRtcConn.Load(cfg.ChannelId)
		if !ok {
			bot.Logger().Info("Channel ID successfully removed from the map", zap.String("channelId", cfg.ChannelId))
		}

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
