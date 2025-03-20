package misskey_test

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	"github.com/wasya-io/petit-misskey/model/misskey"
	"github.com/wasya-io/petit-misskey/service/accounts"
)

func TestMeta(t *testing.T) {
	setting := setting.NewUserSetting()
	service := accounts.NewService(setting)
	account := service.Get("io") // TODO: setupでアカウントを用意する必要がある

	meta := &misskey.Meta{
		AccessToken: account.AccessToken,
		Detail:      false,
	}
	j, err := json.Marshal(meta)
	if err != nil {
		panic(err)
	}
	fmt.Printf("%s\n", j)
}
