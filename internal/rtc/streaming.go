package rtc

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"mezon-go-bot/internal/gst"
	"mezon-go-bot/internal/ws"

	"os"
	"sync"
	"time"

	"github.com/pion/webrtc/v4"
	"github.com/pion/webrtc/v4/pkg/media"
	"github.com/pion/webrtc/v4/pkg/media/oggreader"
)

var (
	MapStreamingRtcConn sync.Map // map[channelId]*RTCConnection
)

type StreamingRTCConn struct {
	peer *webrtc.PeerConnection
	ws   ws.IWSConnection

	clanId      string
	channelId   string
	userId      string
	displayName string

	audioTrack *webrtc.TrackLocalStaticSample
	videoTrack *webrtc.TrackLocalStaticSample
	pipeline   *gst.Pipeline
}

// Stop implements IStreamingRTCConnection.
func (c *StreamingRTCConn) Stop() {
	panic("unimplemented")
}

type IStreamingRTCConnection interface {
	SendAudioTrack(filePath string) error
	Close(channelId string)
	Start()
	Stop()
}

func NewStreamingRTCConnection(config webrtc.Configuration, wsConn ws.IWSConnection, clanId, channelId, userId, displayName string) (IStreamingRTCConnection, error) {
	containerPath := ""
	peerConnection, err := webrtc.NewPeerConnection(config)
	if err != nil {
		return nil, err
	}

	videoTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "video/h264"}, "synced", "video")
	if err != nil {
		log.Fatal("play error")
	}

	audioTrack, err := webrtc.NewTrackLocalStaticSample(webrtc.RTPCodecCapability{MimeType: "audio/opus"}, "synced", "audio")
	if err != nil {
		log.Fatal("play error")
	}

	if _, err := peerConnection.AddTrack(audioTrack); err != nil {
		log.Fatal(err)
	} else if _, err = peerConnection.AddTrack(videoTrack); err != nil {
		log.Fatal(err)
	}

	pipeline := gst.CreatePipeline(containerPath, audioTrack, videoTrack)

	// save to store
	rtcConnection := &StreamingRTCConn{
		peer:        peerConnection,
		ws:          wsConn,
		clanId:      clanId,
		channelId:   channelId,
		userId:      userId,
		displayName: displayName,
		audioTrack:  audioTrack,
		videoTrack:  videoTrack,
		pipeline:    pipeline,
	}

	// ws receive message handler ( on event )
	wsConn.SetOnMessage(rtcConnection.onWebsocketEvent)
	MapStreamingRtcConn.Store(channelId, rtcConnection)

	peerConnection.OnICEConnectionStateChange(func(state webrtc.ICEConnectionState) {
		log.Printf("Connection State has changed %s \n", state.String())

		switch state {
		case webrtc.ICEConnectionStateConnected:
			// TODO: event ice connected
			jsonData, _ := json.Marshal(map[string]string{"ChannelId": channelId})
			wsConn.SendMessage(&ws.WsMsg{
				ClanId:      clanId,
				ChannelId:   channelId,
				Key:         "connect_publisher",
				Value:       jsonData,
				UserId:      userId,
				DisplayName: displayName,
			})
		case webrtc.ICEConnectionStateClosed:
			rtcConn, ok := MapStreamingRtcConn.Load(channelId)
			if !ok {
				return
			}

			if rtcConn.(*StreamingRTCConn).peer == nil {
				return
			}

			if rtcConn.(*StreamingRTCConn).peer.ConnectionState() != webrtc.PeerConnectionStateClosed {
				rtcConn.(*StreamingRTCConn).peer.Close()
			}

			MapStreamingRtcConn.Delete(channelId)
		}
	})
	peerConnection.OnICECandidate(func(i *webrtc.ICECandidate) {
		rtcConnection.onICECandidate(i, channelId, clanId, userId, displayName)
	})

	// send offer
	rtcConnection.sendOffer()

	return rtcConnection, nil
}

func (c *StreamingRTCConn) Start() {
	c.pipeline.Start()
}

func (c *StreamingRTCConn) Close(channelId string) {
	rtcConn, ok := MapStreamingRtcConn.Load(channelId)
	if !ok {
		return
	}

	if rtcConn.(*StreamingRTCConn).peer == nil {
		return
	}

	if rtcConn.(*StreamingRTCConn).peer.ConnectionState() != webrtc.PeerConnectionStateClosed {
		rtcConn.(*StreamingRTCConn).peer.Close()
	}

	MapStreamingRtcConn.Delete(channelId)
}

func (c *StreamingRTCConn) onWebsocketEvent(event *ws.WsMsg) error {

	// TODO: fix hardcode
	switch event.Key {
	case "sd_answer":
		// unzipData, _ := utils.GzipUncompress(eventData.JsonData)
		// var answer webrtc.SessionDescription
		var answerSDP string
		err := json.Unmarshal(event.Value, &answerSDP)
		if err != nil {
			return err
		}

		return c.peer.SetRemoteDescription(webrtc.SessionDescription{
			Type: webrtc.SDPTypeAnswer,
			SDP:  answerSDP,
		})

	case "ice_candidate":

		var i webrtc.ICECandidateInit
		err := json.Unmarshal(event.Value, &i)
		if err != nil {
			return err
		}

		return c.addICECandidate(i)

	case "connect_publisher":
		return c.sendPtt()
	}

	return nil
}

func (c *StreamingRTCConn) sendOffer() error {
	offer, err := c.peer.CreateOffer(nil)
	if err != nil {
		return err
	}
	if err := c.peer.SetLocalDescription(offer); err != nil {
		return err
	}

	byteJson, _ := json.Marshal(offer)
	// dataEnc, _ := utils.GzipCompress(string(byteJson))

	// send socket signaling, gzip compress data
	return c.ws.SendMessage(&ws.WsMsg{
		Key:         "session_publisher",
		ClanId:      c.clanId,
		ChannelId:   c.channelId,
		UserId:      c.userId,
		DisplayName: c.displayName,
		Value:       byteJson,
	})
}

func (c *StreamingRTCConn) sendPtt() error {
	jsonData, _ := json.Marshal(map[string]interface{}{
		"ChannelId": c.channelId,
		"IsTalk":    true,
	})
	return c.ws.SendMessage(&ws.WsMsg{
		Key:         "ptt_publisher",
		ClanId:      c.clanId,
		ChannelId:   c.channelId,
		UserId:      c.userId,
		DisplayName: c.displayName,
		Value:       jsonData,
	})
}

func (c *StreamingRTCConn) onICECandidate(i *webrtc.ICECandidate, clanId, channelId, userId, displayName string) error {
	if i == nil {
		return nil
	}
	// If you are serializing a candidate make sure to use ToJSON
	// Using Marshal will result in errors around `sdpMid`
	candidateString, err := json.Marshal(i.ToJSON())
	if err != nil {
		return err
	}

	return c.ws.SendMessage(&ws.WsMsg{
		Key:         "ice_candidate",
		Value:       candidateString,
		ClanId:      clanId,
		ChannelId:   channelId,
		UserId:      userId,
		DisplayName: displayName,
	})
}

func (c *StreamingRTCConn) addICECandidate(i webrtc.ICECandidateInit) error {
	return c.peer.AddICECandidate(i)
}

func (c *StreamingRTCConn) SendAudioTrack(filePath string) error {

	// Open a OGG file and start reading using our OGGReader
	file, oggErr := os.Open(filePath)
	if oggErr != nil {
		return oggErr
	}

	// Open on oggfile in non-checksum mode.
	ogg, _, oggErr := oggreader.NewWith(file)
	if oggErr != nil {
		return oggErr
	}

	// Keep track of last granule, the difference is the amount of samples in the buffer
	var lastGranule uint64

	// It is important to use a time.Ticker instead of time.Sleep because
	// * avoids accumulating skew, just calling time.Sleep didn't compensate for the time spent parsing the data
	// * works around latency issues with Sleep (see https://github.com/golang/go/issues/44343)
	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()
	for ; true; <-ticker.C {
		pageData, pageHeader, oggErr := ogg.ParseNextPage()
		if errors.Is(oggErr, io.EOF) {
			log.Println("All audio pages parsed and sent")
			return nil
		}

		if oggErr != nil {
			return oggErr
		}

		// The amount of samples is the difference between the last and current timestamp
		sampleCount := float64(pageHeader.GranulePosition - lastGranule)
		lastGranule = pageHeader.GranulePosition
		sampleDuration := time.Duration((sampleCount/48000)*1000) * time.Millisecond

		if oggErr = c.audioTrack.WriteSample(media.Sample{Data: pageData, Duration: sampleDuration}); oggErr != nil {
			return oggErr
		}
	}
	return nil

}
