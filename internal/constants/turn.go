package constants

import "github.com/pion/webrtc/v4"

var ICE_MEZON = webrtc.ICEServer{
	URLs:           []string{"turn:turn.mezon.vn:5349"},
	Username:       "turnmezon",
	Credential:     "QuTs4zUEcbylWemXL7MK",
	CredentialType: webrtc.ICECredentialTypePassword,
}

var ICE_GOOGLE = webrtc.ICEServer{
	URLs: []string{"stun:stun.l.google.com:19302"},
}
