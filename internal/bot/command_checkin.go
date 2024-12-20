package bot

import (
	"errors"
	"ncc/go-mezon-bot/internal/constants"
	"ncc/go-mezon-bot/internal/rtc"
	"ncc/go-mezon-bot/pkg/clients"

	"github.com/pion/webrtc/v4"
	"go.uber.org/zap"
)

func (b *Bot) Checkin() IBot {

	// make call service
	b.callService = rtc.NewCallService(b.botCfg.BotId, b.signalingDM, webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{constants.ICE_MEZON},
	})

	b.callService.SetOnImage(b.checkinHandler, constants.NUM_IMAGE_SNAPSHOT)
	b.callService.SetAcceptCallFileAudio(constants.CHECKIN_ACCEPT_CALL_AUDIO_PATH)
	b.callService.SetExitCallFileAudio(constants.CHECKIN_EXIT_CALL_AUDIO_PATH)
	b.callService.SetCheckinSuccessFileAudio(constants.CHECKIN_CHECKIN_SUCCESS_AUDIO_PATH)
	b.callService.SetCheckinFailFileAudio(constants.CHECKIN_CHECKIN_FAIL_AUDIO_PATH)
	b.signalingDM.SetOnWebrtcSignalingFwd(b.callService.OnWebsocketEvent)

	return b
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
