package bot

import (
	"ncc/go-mezon-bot/internal/constants"
	"ncc/go-mezon-bot/pkg/clients"

	"go.uber.org/zap"
)

func (b *Bot) StartCheckin() {

	b.callService.SetOnImage(b.checkinHandler, constants.NUM_IMAGE_SNAPSHOT)

	b.signaling.SetOnWebrtcSignalingFwd(b.callService.OnWebsocketEvent)

	b.logger.Info("Bot checkin is running...")

}

func (b *Bot) checkinHandler(imageBase64 string) error {
	res, err := clients.CheckinApi(imageBase64)
	if err != nil {
		b.logger.Error("[CheckinApi] error", zap.Error(err))
		return err
	}

	b.logger.Info("[CheckinApi] info", zap.Any("info", res))

	return nil
}
