package providers

const (
	CheckNotFound = int64(-1)

	HeaderAuthorization = "Authorization"
	HeaderAccept        = "Accept"
	HeaderContentType   = "Content-Type"
)

// UptimeProviderID enum of supported providers
type UptimeProviderID string

const (
	ProviderPingdom     UptimeProviderID = "pingdom"
	ProviderBetterStack UptimeProviderID = "betterstack"
	ProviderMock        UptimeProviderID = "mock"
)
