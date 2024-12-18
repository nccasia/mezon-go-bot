package bot

import (
	"errors"
	"ncc/go-mezon-bot/internal/constants"
	"ncc/go-mezon-bot/pkg/clients"

	"go.uber.org/zap"
)

func (b *Bot) StartCheckin() {

	b.callService.SetOnImage(b.checkinHandler, constants.NUM_IMAGE_SNAPSHOT)
	b.callService.SetAcceptCallFileAudio(constants.CHECKIN_ACCEPT_CALL_AUDIO_PATH)
	b.callService.SetExitCallFileAudio(constants.CHECKIN_EXIT_CALL_AUDIO_PATH)
	b.signaling.SetOnWebrtcSignalingFwd(b.callService.OnWebsocketEvent)

	b.logger.Info("Bot checkin is running...")
}

func (b *Bot) checkinHandler(imageBase64 string) error {
	res, err := clients.CheckinApi(imageBase64)
	if err != nil {
		b.logger.Error("[CheckinApi] error", zap.Error(err))
		return err
	}

	b.logger.Info("[CheckinApi] send image", zap.Any("info", res))
	if res.Probability >= constants.CHECKIN_PROBABILITY_SUCCESS {
		b.logger.Info("[CheckinApi] checkin success", zap.Any("info", res))

		// return error close send image base64 to function
		return errors.New("checkin success")
	}

	// return nil -> continue send image base64 to function
	return nil
}
