package nginxproxymanager

import (
	"log"

	"github.com/elouan/dockyard/internal/ports/routing"
)

type NoopProvider struct{}

func (n *NoopProvider) EnsureRoute(req routing.RouteRequest) error {
	log.Printf("npm: noop EnsureRoute hostname=%s forward=%s:%d", req.Hostname, req.ForwardHost, req.TargetPort)
	return nil
}

func (n *NoopProvider) DeleteRoute(hostname string) error {
	log.Printf("npm: noop DeleteRoute hostname=%s", hostname)
	return nil
}
