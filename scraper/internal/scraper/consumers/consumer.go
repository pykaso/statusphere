package consumers

import (
	"github.com/metoro-io/statusphere/common/api"
)

type Consumer interface {
	// Consume consumes the given incidents
	Consume(incidents []api.Incident, scraper string, page api.StatusPage) error
	ConsumeUrl(incidents []api.Incident, scraper string, url string) error
}
