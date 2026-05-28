package nginxproxymanager

type tokenRequest struct {
	Identity string `json:"identity"`
	Secret   string `json:"secret"`
}

type tokenResponse struct {
	Token   string `json:"token"`
	Expires string `json:"expires"`
}

type proxyHost struct {
	ID             int      `json:"id,omitempty"`
	DomainNames    []string `json:"domain_names"`
	ForwardHost    string   `json:"forward_host"`
	ForwardPort    int      `json:"forward_port"`
	ForwardScheme  string   `json:"forward_scheme"`
	SSLForced      bool     `json:"ssl_forced"`
	CachingEnabled bool     `json:"caching_enabled"`
	BlockExploits  bool     `json:"block_exploits"`
}
