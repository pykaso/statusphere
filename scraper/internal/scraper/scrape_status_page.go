package scraper

import (
	"context"

	"github.com/metoro-io/statusphere/common/api"
	"github.com/metoro-io/statusphere/common/utils"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *scraper) ScrapeStatusPageHistorical(ctx context.Context, url string) ([]api.Incident, string, error) {
	ctx = utils.UpdateContextMdc(ctx, map[string]string{"url": url})
	return s.cascadingScrapeHistorical(ctx, url)
}

// cascadingScrapeHistorical is a helper function that will attempt to scrape the status page using a variety of methods
// If one method fails, it will fall back to the next method
// This is useful because different status pages are structured differently
func (s *scraper) cascadingScrapeHistorical(ctx context.Context, url string) ([]api.Incident, string, error) {
	for _, provider := range s.providers {
		ctx = utils.UpdateContextMdc(ctx, map[string]string{"provider": provider.Name()})
		incidents, scraper, err := provider.ScrapeStatusPageHistorical(ctx, url)
		if err == nil {
			utils.GetLogger(ctx, s.logger).Info("Successfully scraped the status page using the provider method")
			return incidents, scraper, nil
		} else {
			utils.GetLogger(ctx, s.logger).Info("Failed to scrape the status page using the provider method", zap.Error(err))
		}
	}
	return nil, "scraper", errors.New("failed to scrape the status page using any of the provider methods")
}

func (s *scraper) ScrapeStatusPageCurrent(ctx context.Context, page api.StatusPage) ([]api.Incident, string, error) {
	ctx = utils.UpdateContextMdc(ctx, map[string]string{"url": page.URL})
	return s.cascadingScrapeCurrent(ctx, page)
}

func (s *scraper) cascadingScrapeCurrent(ctx context.Context, page api.StatusPage) ([]api.Incident, string, error) {
	// Pokud je nastaven PreferredScraper
	if page.PreferredScraper != "" {
		for _, provider := range s.providers {
			if provider.Name() == page.PreferredScraper {
				// Pokusíme se použít preferovaný scraper
				ctx = utils.UpdateContextMdc(ctx, map[string]string{"provider": provider.Name()})
				incidents, scraper, err := provider.ScrapeStatusPageCurrent(ctx, page)
				if err == nil {
					utils.GetLogger(ctx, s.logger).Info("Successfully scraped the status page using the preferred provider")
					return incidents, scraper, nil
				} else {
					utils.GetLogger(ctx, s.logger).Info("Failed to scrape the status page using the preferred provider", zap.Error(err))
				}
			}
		}
	}

	// Původní iterace přes všechny poskytovatele
	for _, provider := range s.providers {
		ctx = utils.UpdateContextMdc(ctx, map[string]string{"provider": provider.Name()})
		incidents, scraper, err := provider.ScrapeStatusPageCurrent(ctx, page)
		if err == nil {
			utils.GetLogger(ctx, s.logger).Info("Successfully scraped the status page using the provider method")
			return incidents, scraper, nil
		} else {
			utils.GetLogger(ctx, s.logger).Info("Failed to scrape the status page using the provider method", zap.Error(err))
		}
	}

	return nil, "scraper", errors.New("failed to scrape the status page using any of the provider methods")
}

// cascadingScrapeCurrent is a helper function that will attempt to scrape the status
// page using a variety of methods
// If one method fails, it will fall back to the next method
// This is useful because different status pages are structured differently
func (s *scraper) cascadingScrapeCurrentOrig(ctx context.Context, page api.StatusPage) ([]api.Incident, string, error) {

	for _, provider := range s.providers {
		ctx = utils.UpdateContextMdc(ctx, map[string]string{"provider": provider.Name()})
		incidents, scraper, err := provider.ScrapeStatusPageCurrent(ctx, page)
		if err == nil {
			utils.GetLogger(ctx, s.logger).Info("Successfully scraped the status page using the provider method")
			return incidents, scraper, nil
		} else {
			utils.GetLogger(ctx, s.logger).Info("Failed to scrape the status page using the provider method", zap.Error(err))
		}
	}
	return nil, "scraper", errors.New("failed to scrape the status page using any of the provider methods")
}
