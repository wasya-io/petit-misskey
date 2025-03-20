package resolver_test

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/wasya-io/petit-misskey/infrastructure/resolver"
)

func TestMisskeyResolver(t *testing.T) {
	resolver := resolver.NewMisskeyStreamUrlResolver()

	url, err := resolver.Resolve(
		"https://misskey.io/api",
		map[string]string{
			"accessToken": "test",
		})
	assert.NoError(t, err)
	assert.Equal(t, "wss://misskey.io/streaming?i=test", url)
}
