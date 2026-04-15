package dns

type RecordRequest struct {
	Hostname string
	Target   string
}

type Provider interface {
	EnsureRecord(request RecordRequest) error
	DeleteRecord(hostname string) error
}
