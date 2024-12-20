package bot

import (
	"ncc/go-mezon-bot/internal/constants"
	radiostation "ncc/go-mezon-bot/internal/radio-station"
	"ncc/go-mezon-bot/internal/rtc"

	"github.com/nccasia/mezon-go-sdk/configs"
	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

func (b *Bot) NCC8() IBot {
	b.commands[constants.NCC8_COMMAND] = b.ncc8Handler
	return b
}

func (b *Bot) ncc8Handler(command string, args []string) error {
	if args[0] == constants.NCC8_ARG_PLAY {
		wsConn, err := radiostation.NewWSConnection(&configs.Config{
			BasePath:     b.cfg.StnDomain,
			Timeout:      15,
			InsecureSkip: b.cfg.InsecureSkip,
			UseSSL:       b.cfg.UseSSL,
		}, b.cfg.ClanId, b.botCfg.ChannelId, b.botCfg.BotId, "NCC8")
		if err != nil {
			b.logger.Error("[ncc8] radiostation new ws signaling error", zap.Error(err))
			return err
		}

		rtcConn, err := rtc.NewStreamingRTCConnection(webrtc.Configuration{
			ICEServers: []webrtc.ICEServer{constants.ICE_GOOGLE},
		}, wsConn, b.cfg.ClanId, b.botCfg.ChannelId, b.botCfg.BotId, "NCC8")
		if err != nil {
			b.logger.Error("[ncc8] new streaming rtc connection error", zap.Error(err))
			return err
		}

		// TODO: get mp3 by args[1]
		// TODO: ffmpeg convert mp3 to ogg: ffmpeg -i test.mp3 -c:a libopus -page_duration 20000 test.ogg
		err = rtcConn.SendAudioTrack("audio/ncc8.ogg")
		if err != nil {
			b.logger.Error("[ncc8] send audio file error", zap.Error(err))
			return err
		}
	}

	return nil
}
