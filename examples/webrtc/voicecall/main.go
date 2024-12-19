package main

import (
	"fmt"
	"ncc/go-mezon-bot/internal/rtc"

	mezonsdk "github.com/nccasia/mezon-go-sdk"
	"github.com/nccasia/mezon-go-sdk/configs"

	"github.com/pion/webrtc/v4"
)

func main() {

	// call
	Call()

	select {}
}

func onImage(b64 string) error {
	fmt.Println(b64)
	return nil
}

func Call() {
	// channelId := "1840660756006178816"
	conn, err := mezonsdk.NewWSConnection(&configs.Config{
		// BasePath: "dev-mezon.nccsoft.vn:7305",
		BasePath:     "api.mezon.vn",
		ApiKey:       "7663586b61444979547175356a5a4d52",
		Timeout:      10,
		InsecureSkip: true,
		UseSSL:       true,
	}, "")
	if err != nil {
		panic(err)
	}
	fmt.Println("ws connected")

	callService := rtc.NewCallService("1835944289075466240", conn, webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{
				URLs:           []string{"turn:turn.mezon.vn:5349"},
				Username:       "turnmezon",
				Credential:     "QuTs4zUEcbylWemXL7MK",
				CredentialType: webrtc.ICECredentialTypePassword,
			},
		},
	})

	callService.SetOnImage(onImage, 10)
	callService.SetAcceptCallFileAudio("../../../audio/hello.ogg")
	callService.SetExitCallFileAudio("../../../audio/exit-call.ogg")

	conn.SetOnWebrtcSignalingFwd(callService.OnWebsocketEvent)

	select {}
}
