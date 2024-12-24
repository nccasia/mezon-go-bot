package main

import (
	"ncc/go-mezon-bot/internal/rtc"

	radiostation "ncc/go-mezon-bot/internal/radio-station"

	"github.com/nccasia/mezon-go-sdk/configs"
	"github.com/pion/webrtc/v4"
)

func main() {

	// streaming
	Streaming()

	select {}
}

func Streaming() {

	clanId := "1779484504377790464"    // KOMU
	channelId := "1838042805130235904" // NCC8-RADIO
	userId := "1826067167154540544"    // KOMU
	displayName := "KOMU"              // longma350
	wsConn, err := radiostation.NewWSConnection(&configs.Config{
		BasePath: "stn.mezon.vn", // prod
		// BasePath: "stn.nccsoft.vn",	// dev
		Timeout:      10,
		InsecureSkip: true,
		// UseSSL:       true,
		UseSSL: false,
	}, clanId, channelId, userId, displayName)
	if err != nil {
		panic(err)
	}

	// ffmpeg -i test.mp3 -c:a libopus -page_duration 20000 test.ogg;
	// filePath := "../../audio/ncc8_example.ogg"
	filePath := "../../audio/lk_thucuoi.ogg"
	rtcConn, err := rtc.NewStreamingRTCConnection(webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs: []string{"stun:stun.l.google.com:19302"}, // TODO: check radio station ice server
			},
		},
	}, wsConn, clanId, channelId, userId, displayName)
	if err != nil {
		panic(err)
	}

	err = rtcConn.SendAudioTrack(filePath)
	if err != nil {
		panic(err)
	}
}
