package bot

import (
	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/rtapi"
)

func (b *Bot) Setup() IBot {
	b.signalingClan.SetOnChannelMessage(func(e *rtapi.Envelope) error {
		return b.handleCommand(e.GetChannelMessage())
	})

	return b
}
