package stdoutconsumer

import (
	"github.com/metoro-io/statusphere/common/api"
	"go.uber.org/zap"
)

type StdoutConsumer struct {
	logger *zap.Logger
}

func NewStdoutConsumer(logger *zap.Logger) *StdoutConsumer {
	return &StdoutConsumer{logger: logger}
}

func (s *StdoutConsumer) Consume(incidents []api.Incident, scraper string, url string) error {
	for _, incident := range incidents {
		s.logger.Info("Incident", zap.Any("incident", incident))
	}
	return nil
}
