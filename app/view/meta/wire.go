//go:build wireinject
// +build wireinject

package meta

import (
	"github.com/google/wire"
	"github.com/wasya-io/petit-misskey/config"
	"github.com/wasya-io/petit-misskey/domain/meta"
	"github.com/wasya-io/petit-misskey/infrastructure/bubbles"
	"github.com/wasya-io/petit-misskey/infrastructure/misskey"
	"github.com/wasya-io/petit-misskey/infrastructure/setting"
	service "github.com/wasya-io/petit-misskey/service/meta"
)

func InitializeModel(instance *setting.Instance) *Model {
	wire.Build(
		NewModel,
		service.NewService,
		config.NewConfig,
		misskey.NewClient,
		bubbles.ProviderSet,
		wire.Bind(new(meta.Client), new(*misskey.Client)),
		wire.Bind(new(bubbles.SimpleViewFactory), new(*bubbles.ViewportFactory)),
	)
	return &Model{}
}
