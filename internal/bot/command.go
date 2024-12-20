package bot

import (
	"encoding/json"
	"ncc/go-mezon-bot/internal/helper"
	"ncc/go-mezon-bot/internal/websocket"

	"github.com/nccasia/mezon-go-sdk/mezon-protobuf/mezon/v2/common/api"
	"go.uber.org/zap"
)

type CommandHandler func(command string, args []string) error

func (b *Bot) handleCommand(msg *api.ChannelMessage) error {

	var msgContent *websocket.MsgContent
	if err := json.Unmarshal([]byte(msg.GetContent()), &msgContent); err != nil || msgContent == nil {
		return err
	}

	command, args := helper.ExtractMessage(msgContent.Content)
	b.logger.Debug("[ExtractMessage]", zap.String("command", command), zap.Any("args", args))
	if handler, exists := b.commands[command]; exists {
		return handler(command, args)
	}
	return nil
}
