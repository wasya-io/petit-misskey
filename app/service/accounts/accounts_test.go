package accounts_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	"github.com/wasya-io/petit-misskey/service/accounts"
)

func TestGetIO(t *testing.T) {
	setting := setting.NewUserSetting()
	accounts := accounts.NewService(setting)

	instance := accounts.Get("io")

	assert.NotNil(t, instance)
}

// TODO: 書き込みのテスト
