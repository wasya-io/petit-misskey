package websocket_test

import (
	"os"
	"testing"

	"github.com/wasya-io/petit-misskey/infrastructure/resolver"
	"github.com/wasya-io/petit-misskey/infrastructure/websocket"
	"github.com/wasya-io/petit-misskey/model/misskey"
	"github.com/wasya-io/petit-misskey/test"
)

func TestGetStream(t *testing.T) {
	cfg := test.NewConfig(t)
	resolver := resolver.NewMisskeyStreamUrlResolver()
	wsClient, _ := websocket.NewClient(cfg.Test.BaseUrl, misskey.AccessToken(cfg.Test.AccessToken), resolver, os.Stdout)
	wsClient.Start()
}
