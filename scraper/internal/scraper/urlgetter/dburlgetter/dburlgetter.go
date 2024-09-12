package dburlgetter

import (
	"context"
	"time"

	"github.com/metoro-io/statusphere/common/api"
	"github.com/metoro-io/statusphere/common/db"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type DBURLGetter struct {
	logger          *zap.Logger
	dbClient        *db.DbClient
	StatusPageCache *cache.Cache
}

func NewDBURLGetter(logger *zap.Logger, client *db.DbClient) *DBURLGetter {
	return &DBURLGetter{
		logger:          logger,
		dbClient:        client,
		StatusPageCache: cache.New(time.Minute*20, time.Minute*10),
	}
}

func (s *DBURLGetter) UpdateLastScrapedTimeHistorical(url string, time time.Time) error {
	statusPage, err := s.dbClient.GetStatusPage(context.Background(), url)
	if err != nil {
		return errors.Wrap(err, "failed to get status page")
	}
	statusPage.LastHistoricallyScraped = time
	err = s.dbClient.UpdateStatusPage(context.Background(), *statusPage)
	if err != nil {
		return errors.Wrap(err, "failed to update status page")
	}
	s.StatusPageCache.Set(url, *statusPage, cache.DefaultExpiration)
	return nil
}

func (s *DBURLGetter) UpdateLastScrapedTime(page api.StatusPage, time time.Time, scraped bool) error {
	statusPage, err := s.dbClient.GetStatusPage(context.Background(), page.URL)
	if err != nil {
		return errors.Wrap(err, "failed to get status page")
	}
	statusPage.LastCurrentlyScraped = time
	if !statusPage.IsIndexed && scraped {
		statusPage.IsIndexed = true
	}
	err = s.dbClient.UpdateStatusPage(context.Background(), *statusPage)
	if err != nil {
		return errors.Wrap(err, "failed to update status page")
	}
	s.StatusPageCache.Set(page.URL, *statusPage, cache.DefaultExpiration)
	return nil
}

const timeToRescrape = 5 * time.Minute

func (s *DBURLGetter) GetUrlsToScrapeOrig() ([]string, error) {
	urlsToUse := []string{}
	items := s.StatusPageCache.Items()
	for k, v := range items {
		statusPage, ok := v.Object.(api.StatusPage)
		if !ok {
			s.logger.Error("failed to cast status page")
			continue
		}
		if time.Since(statusPage.LastCurrentlyScraped) > timeToRescrape {
			urlsToUse = append(urlsToUse, k)
		}
	}
	return urlsToUse, nil
}

func (s *DBURLGetter) GetPagesToScrape() ([]api.StatusPage, error) {
	// urlsToUse := []string{}
	pagesToUse := []api.StatusPage{}

	items := s.StatusPageCache.Items()
	for _, v := range items {
		statusPage, ok := v.Object.(api.StatusPage)
		if !ok {
			s.logger.Error("failed to cast status page")
			continue
		}
		if time.Since(statusPage.LastCurrentlyScraped) > timeToRescrape {
			// urlsToUse = append(urlsToUse, k)
			pagesToUse = append(pagesToUse, statusPage)
		}
	}
	return pagesToUse, nil
}

const timeToRescrapeHistorical = 24 * time.Hour * 7

func (s *DBURLGetter) GetHistoricalUrlsToScrape() ([]string, error) {
	urlsToUse := []string{}
	items := s.StatusPageCache.Items()
	for k, v := range items {
		statusPage, ok := v.Object.(api.StatusPage)
		if !ok {
			s.logger.Error("failed to cast status page")
			continue
		}
		if time.Since(statusPage.LastHistoricallyScraped) > timeToRescrapeHistorical {
			urlsToUse = append(urlsToUse, k)
		}
	}
	return urlsToUse, nil
}

func (s *DBURLGetter) Start() {
	s.UpdateStatusPageCache()
}

func (s *DBURLGetter) UpdateStatusPageCache() {
	// Update the cache every 1 minute
	s.updateStatusPageCacheInner()
	ticker := time.NewTicker(1 * time.Minute)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.updateStatusPageCacheInner()
			}
		}
	}()
}

func (s *DBURLGetter) updateStatusPageCacheInner() {
	s.logger.Info("updating status page cache")
	statusPages, err := s.dbClient.GetAllStatusPages(context.Background())
	if err != nil {
		s.logger.Error("failed to get status pages", zap.Error(err))
		return
	}
	for _, statusPage := range statusPages {
		s.StatusPageCache.Set(statusPage.URL, statusPage, cache.DefaultExpiration)
	}
}
