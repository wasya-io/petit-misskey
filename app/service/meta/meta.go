package meta

import (
	"context"

	"github.com/wasya-io/petit-misskey/domain/meta"
	model "github.com/wasya-io/petit-misskey/model/misskey"
)

type Service struct {
	client      meta.Client
	accessToken model.AccessToken
}

func NewService(client meta.Client) *Service {
	return &Service{
		client: client,
	}
}

func (s *Service) Do(ctx context.Context) (*model.MetaResponse, error) {
	contents := &model.Meta{
		AccessToken: s.accessToken,
		Detail:      false,
	}
	res, clientErr := s.client.Meta(ctx, *contents)
	if clientErr != nil {
		return nil, clientErr
	}
	return res, nil
}
