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

	clanId := "1775732550744936448"    // KOMU
	channelId := "1837040466697129984" // NCC8
	userId := "1835944289075466240"    // komu
	// clanId := "3456110592"    // KOMU_2
	// channelId := "4311748608" // NCC8
	// userId := "4198400"       // longma350
	displayName := "KOMU" // longma350
	wsConn, err := radiostation.NewWSConnection(&configs.Config{
		BasePath: "stn.nccsoft.vn",
		// BasePath:     "localhost:8081",
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
