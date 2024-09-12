package main

import (
	"context"
	"net/http"

	"github.com/metoro-io/statusphere/common/db"
	"github.com/metoro-io/statusphere/scraper/internal/scraper"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/consumers"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/consumers/dbconsumer"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/dbgroomer"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/poller"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers/atlassian"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers/rest"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers/rss"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers/rss_ckp"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/urlgetter/dburlgetter"
	"go.uber.org/zap"
)

func main() {
	logger, err := zap.NewProduction()
	if err != nil {
		panic(err)
	}

	scraper := scraper.NewScraper(logger, http.DefaultClient, []providers.Provider{
		atlassian.NewAtlassianProvider(logger, http.DefaultClient),
		rss.NewRssProvider(logger, http.DefaultClient),
		rss_ckp.NewCkpRssProvider(logger, http.DefaultClient),
		rest.NewRestProvider(logger, http.DefaultClient),
	})

	dbClient, err := db.NewDbClientFromEnvironment(logger)
	if err != nil {
		logger.Error("failed to create db client", zap.Error(err))
		return
	}

	err = dbClient.AutoMigrate(context.Background())
	if err != nil {
		logger.Error("failed to auto migrate", zap.Error(err))
		return
	}

	getter := dburlgetter.NewDBURLGetter(logger, dbClient)
	getter.Start()
	dbGroomer := dbgroomer.NewDbGroomer(logger, dbClient)
	dbGroomer.Groom()
	poller := poller.NewPoller(getter, scraper, []consumers.Consumer{
		dbconsumer.NewDbConsumer(logger, dbClient),
	}, logger)
	err = poller.Poll()
	if err != nil {
		logger.Error("failed to poll", zap.Error(err))
		return
	}
}
