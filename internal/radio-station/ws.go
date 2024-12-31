package radiostation

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/nccasia/mezon-go-sdk/configs"
	"github.com/nccasia/mezon-go-sdk/utils"
)

func recvDefaultHandler(e *WsMsg) error {
	return nil
}

type WSConnection struct {
	conn      *websocket.Conn
	dialer    *websocket.Dialer
	basePath  string
	token     string
	clanId    string
	channelId string
	userId    string
	username  string
	mu        sync.Mutex
	onMessage func(*WsMsg) error
}

type IWSConnection interface {
	SetOnMessage(recvHandler func(*WsMsg) error)
	SendMessage(data *WsMsg) error
}

// TODO: implement (TODO) for IWSConnection

func NewWSConnection(c *configs.Config, clanId, channelId, userId, username, token string) (IWSConnection, error) {

	// TODO: authenticate token for ws
	// token, err := getAuthenticate(c)
	// if err != nil {
	// 	return nil, err
	// }

	client := &WSConnection{
		username: username,
		token:    token,
		basePath: utils.GetBasePath("wss", c.BasePath, c.UseSSL),
		// basePath:  utils.GetBasePath("ws", c.BasePath, c.UseSSL),
		clanId:    clanId,
		onMessage: recvDefaultHandler,
	}

	if c.InsecureSkip {
		tlsConfig := &tls.Config{
			InsecureSkipVerify: true,
		}
		client.dialer = &websocket.Dialer{
			TLSClientConfig: tlsConfig,
		}
	} else {
		client.dialer = websocket.DefaultDialer
	}

	if err := client.newWSConnection(); err != nil {
		return nil, err
	}

	return client, nil
}

func (s *WSConnection) newWSConnection() error {

	// TODO: authenticate token for ws
	conn, _, err := s.dialer.Dial(fmt.Sprintf("%s/ws?username=%s&token=%s", s.basePath, s.username, s.token), nil)
	if err != nil {
		log.Println("WebSocket connection open err: ", err)
		return err
	}

	s.conn = conn

	s.reconnect()
	s.pingPong()
	s.recvMessage()

	return nil
}

func (s *WSConnection) Close() error {
	return s.conn.Close()
}

func (s *WSConnection) SendMessage(data *WsMsg) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	jsonData, err := json.Marshal(data)
	if err != nil {
		return err
	}

	return s.conn.WriteMessage(websocket.TextMessage, jsonData)
}

func (s *WSConnection) pingPong() {
	// Ping Handler
	// TODO:
}

func (s *WSConnection) reconnect() {
	go func() {
		for {
			_, _, err := s.conn.ReadMessage()
			if err != nil {
				log.Println("WebSocket disconnected, attempting to reconnect: ", err)

				_ = s.conn.Close()

				for {
					conn, _, err := s.dialer.Dial(fmt.Sprintf("%s/ws?username=%s&token=%s", s.basePath, s.username, s.token), nil)
					if err != nil {
						log.Println("Reconnect failed, retrying in 5 seconds: ", err)
						time.Sleep(5 * time.Second)
						continue
					}

					s.conn = conn
					log.Println("WebSocket reconnected successfully")
					break
				}

				go s.recvMessage()
				go s.pingPong()
			}
		}
	}()
}

func (s *WSConnection) recvMessage() {
	go func() {
		for {
			msgType, databytes, err := s.conn.ReadMessage()
			if err != nil {
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway) ||
					websocket.IsUnexpectedCloseError(err) {
					log.Println("WebSocket connection closed:", err)
					return
				}
				continue
			}

			if msgType != websocket.TextMessage {
				log.Println("unknown message type: ", msgType)
				continue
			}

			var msg WsMsg
			err = json.Unmarshal(databytes, &msg)
			if err != nil {
				log.Println("can't unmarshal json data: ", string(databytes))
				continue
			}

			if err := s.onMessage(&msg); err != nil {
				log.Println("on message error: ", err.Error())
				continue
			}
		}
	}()
}

func (s *WSConnection) SetOnMessage(recvHandler func(*WsMsg) error) {
	s.onMessage = recvHandler
}
