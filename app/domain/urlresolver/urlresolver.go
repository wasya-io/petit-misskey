package urlresolver

type (
	Resolver interface {
		Resolve(baseUrl string, params map[string]string) (string, error)
	}
)
