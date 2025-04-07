package misskey_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/wasya-io/petit-misskey/infrastructure/misskey"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	model "github.com/wasya-io/petit-misskey/model/misskey"
	"github.com/wasya-io/petit-misskey/test"
	"github.com/wasya-io/petit-misskey/util"
)

func TestMeta(t *testing.T) {
	config := test.NewConfig(t)
	setting := setting.NewUserSetting()
	instance := setting.GetInstanceByKey("io")

	client := misskey.NewClient(
		config,
		instance,
	)
	body := &model.Meta{
		Detail: false,
	}
	result, err := client.Meta(context.Background(), *body)
	if err != nil {
		fmt.Printf("%v", err)
	}
	fmt.Println(util.PrittyJson(result))
}
