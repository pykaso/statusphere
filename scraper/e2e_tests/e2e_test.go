package e2e_tests

import (
	"context"
	"net/http"
	"testing"

	"github.com/metoro-io/statusphere/common/api"
	"github.com/metoro-io/statusphere/common/status_pages"
	"github.com/metoro-io/statusphere/scraper/internal/scraper"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers/atlassian"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers/rest"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers/rss_ckp"
	"go.uber.org/zap"
)

var statusPages = []string{
	"https://status.dropbox.com",
	"https://www.calendlystatus.com",
	"https://status.whatnot.com",
	"https://www.githubstatus.com",
	"https://status.multiversx.com",
	"https://status.1password.com",
	"https://status.edq.com",
	"https://status.snowflake.com",
	"https://status.redhat.com",
	"https://status.payscale.com",
}

func TestE2eDropboxHistorical(t *testing.T) {
	dev, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("Failed to create logger")
	}
	scraper := scraper.NewScraper(dev, http.DefaultClient, []providers.Provider{atlassian.NewAtlassianProvider(dev, http.DefaultClient)})
	status_page := api.StatusPage{
		URL: "https://status.dropbox.com",
	}
	incidents, _, err := scraper.ScrapeStatusPageHistorical(context.Background(), status_page.URL)
	if err != nil {
		t.Errorf("Failed to scrape status page: %s", "https://status.dropbox.com")
	}
	dev.Info("Incidents from status page", zap.Any("statusPage", "https://status.dropbox.com"), zap.Any("numIncidents", len(incidents)))
}

func TestE2eCloudflareCurrent(t *testing.T) {
	dev, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("Failed to create logger")
	}
	scraper := scraper.NewScraper(dev, http.DefaultClient, []providers.Provider{atlassian.NewAtlassianProvider(dev, http.DefaultClient)})
	status_page := api.StatusPage{
		URL: "https://www.cloudflarestatus.com",
	}
	incident, _, err := scraper.ScrapeStatusPageCurrent(context.Background(), status_page)
	if err != nil {
		t.Errorf("Failed to scrape status page: %s", "https://www.cloudflarestatus.com")
	}
	dev.Info("Incidents from status page", zap.Any("statusPage", "https://www.cloudflarestatus.com"), zap.Any("numIncidents", len(incident)))
}

func TestE2eManyCurrent(t *testing.T) {
	dev, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("Failed to create logger")
	}
	scraper := scraper.NewScraper(dev, http.DefaultClient, []providers.Provider{atlassian.NewAtlassianProvider(dev, http.DefaultClient)})
	for _, url := range statusPages {

		status_page := api.StatusPage{
			URL: url,
		}
		incidents, _, err := scraper.ScrapeStatusPageCurrent(context.Background(), status_page)
		if err != nil {
			t.Errorf("Failed to scrape status page: %s", status_page.URL)
		}
		dev.Info("Incidents from status page", zap.Any("statusPage", status_page), zap.Any("numIncidents", len(incidents)))
	}
}

func TestE2eCkpRSS(t *testing.T) {
	dev, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("Failed to create logger")
	}
	scraper := scraper.NewScraper(dev, http.DefaultClient, []providers.Provider{rss_ckp.NewCkpRssProvider(dev, http.DefaultClient)})
	var statusPage = status_pages.PageCKP
	incidents, _, err := scraper.ScrapeStatusPageCurrent(context.Background(), statusPage)
	if err != nil {
		t.Errorf("Failed to scrape status page: %s", statusPage.URL)
	}
	dev.Info("Incidents from status page", zap.Any("statusPage", statusPage), zap.Any("numIncidents", len(incidents)))

}

func TestE2eRestApiInstacover(t *testing.T) {
	dev, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("Failed to create logger")
	}
	scraper := scraper.NewScraper(dev, http.DefaultClient, []providers.Provider{rest.NewRestProvider(dev, http.DefaultClient)})

	var statusPage = status_pages.PageInstacover
	incidents, _, err := scraper.ScrapeStatusPageCurrent(context.Background(), statusPage)
	if err != nil {
		t.Error(err)
		t.Errorf("Failed to scrape status page: %s.", statusPage.URL)
	}
	dev.Info("Incidents from status page", zap.Any("statusPage", statusPage), zap.Any("numIncidents", len(incidents)))
}

func TestE2eRestApiSmartform(t *testing.T) {
	dev, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("Failed to create logger")
	}
	scraper := scraper.NewScraper(dev, http.DefaultClient, []providers.Provider{rest.NewRestProvider(dev, http.DefaultClient)})

	var statusPage = status_pages.PageSmartform
	incidents, _, err := scraper.ScrapeStatusPageCurrent(context.Background(), statusPage)
	if err != nil {
		t.Error(err)
		t.Errorf("Failed to scrape status page: %s.", statusPage.URL)
	}
	dev.Info("Incidents from status page", zap.Any("statusPage", statusPage), zap.Any("numIncidents", len(incidents)))
}

func TestE2eRestApiAres(t *testing.T) {
	dev, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("Failed to create logger")
	}
	scraper := scraper.NewScraper(dev, http.DefaultClient, []providers.Provider{rest.NewRestProvider(dev, http.DefaultClient)})

	var statusPage = status_pages.PageAres
	incidents, _, err := scraper.ScrapeStatusPageCurrent(context.Background(), statusPage)
	if err != nil {
		t.Error(err)
		t.Errorf("Failed to scrape status page: %s.", statusPage.URL)
	}
	dev.Info("Incidents from status page", zap.Any("statusPage", statusPage), zap.Any("numIncidents", len(incidents)))
}

func TestE2eRestIPEX(t *testing.T) {
	dev, err := zap.NewDevelopment()
	if err != nil {
		t.Errorf("Failed to create logger")
	}
	scraper := scraper.NewScraper(dev, http.DefaultClient, []providers.Provider{rest.NewRestProvider(dev, http.DefaultClient)})

	var statusPage = status_pages.PageIPEX
	incidents, _, err := scraper.ScrapeStatusPageCurrent(context.Background(), statusPage)
	if err != nil {
		t.Error(err)
		t.Errorf("Failed to scrape status page: %s.", statusPage.URL)
	}
	dev.Info("Incidents from status page", zap.Any("statusPage", statusPage), zap.Any("numIncidents", len(incidents)))
}
