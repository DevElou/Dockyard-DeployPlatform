package routing

type RouteRequest struct {
	Hostname      string
	ForwardHost   string // actual host NPM should proxy to
	TargetPort    int
	ForwardScheme string // http or https; empty uses provider default
	TLS           bool
}

type Provider interface {
	EnsureRoute(request RouteRequest) error
	DeleteRoute(hostname string) error
}
