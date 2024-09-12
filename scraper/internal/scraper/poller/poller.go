package poller

import (
	"context"
	"time"

	"github.com/metoro-io/statusphere/common/api"
	"github.com/metoro-io/statusphere/scraper/internal/scraper"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/consumers"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/urlgetter"
	"github.com/patrickmn/go-cache"
	"go.uber.org/zap"
)

type Poller struct {
	urlGetter                           urlgetter.URLGetter
	scraper                             scraper.Scraper
	consumers                           []consumers.Consumer
	currentlyExecutingScrapes           *cache.Cache
	currentlyExecutingHistoricalScrapes *cache.Cache
	logger                              *zap.Logger
}

func NewPoller(urlGetter urlgetter.URLGetter, scraper scraper.Scraper, consumers []consumers.Consumer, logger *zap.Logger) *Poller {
	return &Poller{
		urlGetter:                           urlGetter,
		scraper:                             scraper,
		consumers:                           consumers,
		currentlyExecutingScrapes:           cache.New(cache.NoExpiration, cache.NoExpiration),
		currentlyExecutingHistoricalScrapes: cache.New(cache.NoExpiration, cache.NoExpiration),
		logger:                              logger,
	}
}

// Poll polls the scraper and sends the incidents to the consumers
// It blocks forever unless an unrecoverable error occurs
func (p *Poller) Poll() error {
	ticker := time.NewTicker(1 * time.Second)
	for {
		select {
		case <-ticker.C:
			err := p.pollInner()
			if err != nil {
				p.logger.Error("failed to poll", zap.Error(err))
			}
			err = p.pollInnerHistorical()
			if err != nil {
				p.logger.Error("failed to poll", zap.Error(err))
			}
		}
	}
}

func (p *Poller) pollInner() error {
	//urlsToScrape, err := p.urlGetter.GetUrlsToScrape()
	pagesToScrape, err := p.urlGetter.GetPagesToScrape()
	if err != nil {
		return err
	}

	var urlsToScrapeWhichAreNotCurrentlyExecuting []api.StatusPage
	for _, page := range pagesToScrape {
		if _, found := p.currentlyExecutingScrapes.Get(page.URL); !found {
			urlsToScrapeWhichAreNotCurrentlyExecuting = append(urlsToScrapeWhichAreNotCurrentlyExecuting, page)
		}
	}

	for _, page := range urlsToScrapeWhichAreNotCurrentlyExecuting {
		go func(page api.StatusPage) {
			p.logger.Info("scraping", zap.String("url", page.URL))
			defer p.logger.Info("finished scraping", zap.String("url", page.URL))
			p.currentlyExecutingScrapes.Set(page.URL, true, cache.NoExpiration)
			defer p.currentlyExecutingScrapes.Delete(page.URL)
			err := p.executeScrape(page)
			successfullyScraped := err == nil
			defer func(urlGetter urlgetter.URLGetter, page api.StatusPage, time time.Time) {
				_ = urlGetter.UpdateLastScrapedTime(page, time, successfullyScraped)
			}(p.urlGetter, page, time.Now())
			if err != nil {
				p.logger.Error("failed to scrape", zap.Error(err), zap.String("url", page.URL))
				return
			}
		}(page)
	}
	return nil
}

// func (p *Poller) executeScrape(url string) error {
func (p *Poller) executeScrape(page api.StatusPage) error {
	incidents, scraper, err := p.scraper.ScrapeStatusPageCurrent(context.Background(), page)
	if err != nil {
		return err
	}
	for _, consumer := range p.consumers {
		err := consumer.Consume(incidents, scraper, page)
		if err != nil {
			return err
		}
	}
	return nil
}

func (p *Poller) pollInnerHistorical() error {
	urlsToScrape, err := p.urlGetter.GetHistoricalUrlsToScrape()
	if err != nil {
		return err
	}

	var urlsToScrapeWhichAreNotCurrentlyExecuting []string
	for _, url := range urlsToScrape {
		if _, found := p.currentlyExecutingHistoricalScrapes.Get(url); !found {
			urlsToScrapeWhichAreNotCurrentlyExecuting = append(urlsToScrapeWhichAreNotCurrentlyExecuting, url)
		}
	}

	for _, url := range urlsToScrapeWhichAreNotCurrentlyExecuting {
		go func(url string) {
			p.logger.Info("scraping historical", zap.String("url", url))
			defer p.logger.Info("finished scraping historical", zap.String("url", url))
			p.currentlyExecutingHistoricalScrapes.Set(url, true, cache.NoExpiration)
			defer func(urlGetter urlgetter.URLGetter, url string, time time.Time) {
				_ = urlGetter.UpdateLastScrapedTimeHistorical(url, time)
			}(p.urlGetter, url, time.Now())
			defer p.currentlyExecutingHistoricalScrapes.Delete(url)
			err := p.executeScrapeHistorical(url)
			if err != nil {
				p.logger.Error("failed to scrape historical", zap.Error(err), zap.String("url", url))
			}
		}(url)
	}
	return nil
}

func (p *Poller) executeScrapeHistorical(url string) error {
	p.currentlyExecutingHistoricalScrapes.Set(url, struct{}{}, cache.NoExpiration)
	incidents, scraper, err := p.scraper.ScrapeStatusPageHistorical(context.Background(), url)
	if err != nil {
		return err
	}
	for _, consumer := range p.consumers {
		err := consumer.ConsumeUrl(incidents, scraper, url)
		if err != nil {
			return err
		}
	}
	return nil
}
