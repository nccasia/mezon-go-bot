package ws

import "encoding/json"

type MsgContent struct {
	Content string `json:"t"`
}

type WsMsg struct {
	Key         string
	ClanId      string
	ChannelId   string
	UserId      string
	DisplayName string
	Value       json.RawMessage
}
