package bot

import (
	"ncc/go-mezon-bot/internal/constants"
	"ncc/go-mezon-bot/pkg/clients"

	"go.uber.org/zap"
)

func (b *Bot) StartCheckin() {

	b.callService.SetOnImage(b.putImage, constants.NUM_IMAGE_SNAPSHOT)
	b.callService.SetFileAudio(constants.CHECKIN_AUDIO_PATH)
	b.signaling.SetOnWebrtcSignalingFwd(b.callService.OnWebsocketEvent)

	b.start()

	b.logger.Info("Bot checkin is running...")
}

func (b *Bot) putImage(imageBase64 string) error {
	b.checkinChan <- imageBase64
	return nil
}

func (b *Bot) checkinHandler(imageBase64 string) {
	res, err := clients.CheckinApi(imageBase64)
	if err != nil {
		b.logger.Error("[CheckinApi] error", zap.Error(err))
		return
	}

	b.logger.Info("[CheckinApi] info", zap.Any("info", res))
}

func (b *Bot) start() {
	go func() {
		for {
			select {
			case data := <-b.checkinChan:
				b.checkinHandler(data)
			}
		}
	}()
}
