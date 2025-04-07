package api

import (
	"context"

	"github.com/wasya-io/petit-misskey/model/misskey"
)

type (
	Client interface {
		CreateNote(ctx context.Context, visibility misskey.Visibility, text string) (*misskey.CreateNoteResponse, error)
	}
)
