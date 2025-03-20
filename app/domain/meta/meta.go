package meta

import (
	"context"

	"github.com/wasya-io/petit-misskey/model/misskey"
)

type (
	Client interface {
		Meta(ctx context.Context, contents misskey.Meta) (*misskey.MetaResponse, error)
	}
)
