package websocket

import (
	"context"
	_ "embed"
	"encoding/json"
	"fmt"
	"io"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/google/uuid"
	"github.com/google/wire"
	"github.com/sacOO7/gowebsocket"
	"github.com/wasya-io/petit-misskey/domain/core"
	"github.com/wasya-io/petit-misskey/domain/urlresolver"
	"github.com/wasya-io/petit-misskey/infrastructure/resolver"
	"github.com/wasya-io/petit-misskey/model/misskey"
)

type (
	Client interface {
		Start() error
		Stop()
		SetWriter(w io.Writer)
		SetTimeline(timelineType string) error
		Pong()
	}

	StandardClient struct {
		baseUrl     string
		accessToken string
		urlResolver urlresolver.Resolver
		writer      io.Writer
		msgCh       chan tea.Msg
		ctx         context.Context
		cancel      context.CancelFunc
		socket      *gowebsocket.Socket
		logger      core.Logger
	}
	ConnectChannelPayload struct {
		Type string      `json:"type"`
		Body PayloadBody `json:"body"`
	}
	PayloadBody struct {
		Channel string `json:"channel,omitempty"`
		Id      string `json:"id"`
	}

	NoteMessage struct {
		Note *misskey.Note `json:"note"`
	}

	WebSocketErrorMsg struct {
		Err error `json:"error"`
	}

	WebSocketConnectedMsg struct{}

	WebSocketDisconnectedMsg struct {
		Err error `json:"error"`
	}

	WebSocketPingReceivedMsg struct {
		Data string `json:"data"`
	}
)

var (
	ChannelTypeMain  = "main"
	ChannelTypeHome  = "homeTimeline"
	ChannelTypeLocal = "localTimeline"
)

var ProviderSet = wire.NewSet(
	NewClient,
	wire.Bind(new(urlresolver.Resolver), new(*resolver.MisskeyStreamUrlResolver)), // FIXME: bindはここじゃなくて利用側(usecase層)に書く
)

func NewClient(baseUrl string, accessToken misskey.AccessToken, urlResolver urlresolver.Resolver, writeTo io.Writer, logger core.Logger) (Client, chan tea.Msg) {
	ctx, cancel := context.WithCancel(context.Background())

	msgCh := make(chan tea.Msg, 100)
	return &StandardClient{
		baseUrl:     baseUrl,
		accessToken: string(accessToken),
		urlResolver: urlResolver,
		writer:      nil,
		msgCh:       msgCh,
		ctx:         ctx,
		cancel:      cancel,
		logger:      logger,
	}, msgCh
}

func (c *StandardClient) SetWriter(w io.Writer) {
	c.writer = w
}

func (c *StandardClient) Start() error {
	wsUrl, resolveErr := c.urlResolver.Resolve(
		c.baseUrl,
		map[string]string{
			"accessToken": c.accessToken,
		})
	if resolveErr != nil {
		return resolveErr
	}

	socket := gowebsocket.New(wsUrl)
	c.socket = &socket

	c.socket.OnConnected = func(socket gowebsocket.Socket) {
		c.logger.Log("websocket", "Connected to WebSocket server")
		if c.msgCh != nil {
			c.msgCh <- WebSocketConnectedMsg{}
		}
	}

	c.socket.OnConnectError = func(err error, socket gowebsocket.Socket) {
		c.logger.Log("websocket", fmt.Sprintf("WebSocket connection error: %v", err))
		if c.msgCh != nil {
			c.msgCh <- WebSocketErrorMsg{Err: err}
		}
	}

	c.socket.OnTextMessage = func(message string, socket gowebsocket.Socket) {
		// TODO: このあたりの描画処理はまるごとwriterへ委譲する
		c.logger.Log("websocket", fmt.Sprintf("Received message: %s", message))
		note := &misskey.Note{}
		if err := json.Unmarshal([]byte(message), &note); err != nil {
			// log.Printf("note marshalize error %v", err)
			return
		}

		if c.msgCh != nil {
			c.msgCh <- NoteMessage{Note: note}
		}
	}

	c.socket.OnDisconnected = func(err error, socket gowebsocket.Socket) {
		c.logger.Log("websocket", fmt.Sprintf("Disconnected from WebSocket server: %v", err))
		if c.msgCh != nil {
			c.msgCh <- WebSocketDisconnectedMsg{Err: err}
		}
	}

	c.socket.OnPingReceived = func(data string, socket gowebsocket.Socket) {
		c.logger.Log("websocket", fmt.Sprintf("Ping received: %s", data))
		if c.msgCh != nil {
			c.msgCh <- WebSocketPingReceivedMsg{Data: data}
		}
		c.Pong()
	}

	c.socket.OnPongReceived = func(data string, socket gowebsocket.Socket) {
		c.logger.Log("websocket", fmt.Sprintf("Pong received: %s", data))
		if c.msgCh != nil {
			c.msgCh <- WebSocketPingReceivedMsg{Data: data}
		}
	}

	c.socket.Connect()

	uu, err := uuid.NewRandom()
	if err != nil {
		return err
	}
	tlChId := uu.String()

	connectLocalBody := &PayloadBody{
		Channel: ChannelTypeHome,
		Id:      tlChId,
	}
	homeText, _ := json.Marshal(&ConnectChannelPayload{Type: "connect", Body: *connectLocalBody})
	c.socket.SendText(string(homeText))

	<-c.ctx.Done()

	// 接続を閉じる準備
	log.Println("Closing WebSocket connection...")

	disconnectBody := &PayloadBody{
		Id: tlChId,
	}
	disconnectText, _ := json.Marshal(&ConnectChannelPayload{Type: "disconnect", Body: *disconnectBody})
	c.socket.SendText(string(disconnectText))
	c.socket.Close()
	c.logger.Flush()
	return nil
}

// Stop はWebSocket接続を終了します
func (c *StandardClient) Stop() {
	c.cancel()
}

func (c *StandardClient) Pong() {
	c.socket.SendText("pong")
}

// SetTimeline はタイムラインの種類を変更します
func (c *StandardClient) SetTimeline(timelineType string) error {
	// if c.socket == nil {
	// 	return errors.New("WebSocket接続が確立されていません")
	// }

	// 既存のタイムラインから切断するコードが必要
	// ...

	// 新しいタイムラインに接続
	uu, err := uuid.NewRandom()
	if err != nil {
		return fmt.Errorf("チャネルID生成エラー: %w", err)
	}
	chId := uu.String()

	connectBody := &PayloadBody{
		Channel: timelineType,
		Id:      chId,
	}
	connectText, _ := json.Marshal(&ConnectChannelPayload{Type: "connect", Body: *connectBody})
	c.socket.SendText(string(connectText))

	return nil
}
