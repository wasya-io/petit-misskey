package resolver

import (
	"fmt"
	"net/url"

	"github.com/pkg/errors"
)

type (
	MisskeyStreamUrlResolver struct{}
)

func NewMisskeyStreamUrlResolver() *MisskeyStreamUrlResolver {
	return &MisskeyStreamUrlResolver{}
}

func (r *MisskeyStreamUrlResolver) Resolve(baseUrl string, params map[string]string) (string, error) {
	urlInfo, err := url.Parse(baseUrl)
	if err != nil {
		return "", errors.WithStack(err) // TODO: infra層から戻すerror用のinterfaceを作るか？
	}
	var accessToken string
	var ok bool
	if accessToken, ok = params["accessToken"]; !ok {
		return "", errors.New("parameter accessToken not found")
	}
	wsUrl := fmt.Sprintf("wss://%s/streaming?i=%s", urlInfo.Host, accessToken)

	return wsUrl, nil
}
