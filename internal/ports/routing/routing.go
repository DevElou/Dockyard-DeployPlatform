package routing

type RouteRequest struct {
	Hostname          string
	TargetServiceName string
	TargetPort        int
	TLS               bool
}

type Provider interface {
	EnsureRoute(request RouteRequest) error
	DeleteRoute(hostname string) error
}
