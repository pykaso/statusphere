package rss_ckp

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"github.com/metoro-io/statusphere/common/api"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

func (s *CkpRssProvider) Name() string {
	return string(providers.ProviderCKP)
}

type CkpRssProvider struct {
	logger     *zap.Logger
	httpClient *http.Client
}

func NewCkpRssProvider(logger *zap.Logger, httpClient *http.Client) *CkpRssProvider {
	return &CkpRssProvider{
		logger:     logger,
		httpClient: httpClient,
	}
}

func (s *CkpRssProvider) ScrapeStatusPageCurrent(ctx context.Context, page api.StatusPage) ([]api.Incident, string, error) {
	return s.scrapeRssPage(ctx, page)
}

func (s *CkpRssProvider) ScrapeStatusPageHistorical(ctx context.Context, url string) ([]api.Incident, string, error) {
	var incidents []api.Incident
	return incidents, s.Name(), nil
}

// scrapeRssPage is a helper function that will attempt to scrape the status
// page using the rss method
// If the ress method fails, it will return an error
func (s *CkpRssProvider) scrapeRssPage(ctx context.Context, page api.StatusPage) ([]api.Incident, string, error) {
	rssPage, isRssPage, err := s.isRssPage(page.URL)
	if err != nil {
		return nil, s.Name(), errors.Wrap(err, "failed to determine if the page is an rss page")
	}
	if !isRssPage {
		return nil, s.Name(), errors.New("page is not a rss page")
	}

	print(rssPage)

	return s.getIncidentsFromRssPage(rssPage, page.URL)
}

// We determine if a page is an rss page by checking if there is a /history page and
// that history page contains the data-react-class='HistoryIndex' attribute
func (s *CkpRssProvider) isRssPage(url string) (string, bool, error) {
	response, err := s.httpClient.Get(url)
	if err != nil {
		return "", false, errors.Wrap(err, "failed to make the get request to the history page")
	}
	if response.StatusCode != http.StatusOK {
		return "", false, fmt.Errorf("Invalid response StatusCode: %w", response.StatusCode)
	}

	// Is the body well formed xml?
	if isXMLContent(response.Body) {
		return url, true, nil
	}
	return "", false, nil
}

// isXMLContent tries to parse the response body as XML to check if it is valid XML.
func isXMLContent(body io.Reader) bool {
	decoder := xml.NewDecoder(body)
	for {
		token, err := decoder.Token()
		if err != nil {
			// Return false if there is an error during decoding (e.g., invalid XML)
			return false
		}

		switch elem := token.(type) {
		case xml.StartElement:
			// Check if the root element is <rss>
			if strings.ToLower(elem.Name.Local) == "rss" {
				// If we find <rss> as the root element, it's likely an RSS feed
				return true
			}
		case xml.Directive:
			// Check if the directive contains "html", which might indicate it's not XML content
			if strings.Contains(strings.ToLower(string(elem)), "html") {
				return false
			}
		}
	}
}

// RssItem represents a single RSS item
type RssItem struct {
	Title       string `xml:"title"`
	Description string `xml:"description"`
	Guid        string `xml:"guid"`
	PubDate     string `xml:"pubDate"`
}

// RssChannel represents the RSS channel that contains multiple items
type RssChannel struct {
	Title       string    `xml:"title"`
	Link        string    `xml:"link"`
	Description string    `xml:"description"`
	Items       []RssItem `xml:"item"`
}

// RssFeed represents the RSS feed structure
type RssFeed struct {
	Channel RssChannel `xml:"channel"`
}

func (s *CkpRssProvider) getIncidentsFromRssPage(url string, statusPageUrl string) ([]api.Incident, string, error) {
	var incidents []api.Incident
	currentYear := time.Now().Year()

	// Fetch the RSS feed
	resp, err := http.Get(url)
	if err != nil {
		return nil, s.Name(), fmt.Errorf("failed to fetch the feed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, s.Name(), fmt.Errorf("failed to read the feed: %w", err)
	}

	// Parse the RSS feed
	var rssFeed RssFeed
	err = xml.Unmarshal(body, &rssFeed)
	if err != nil {
		return nil, s.Name(), fmt.Errorf("failed to parse the feed: %w", err)
	}

	// Regex patterns for extracting date and time
	reSingleDay := regexp.MustCompile(`od (\d{2}\.\d{2}\.\d{4} \d{2}:\d{2}) do (\d{2}:\d{2})`)
	reMultiDay := regexp.MustCompile(`od (\d{2}\.\d{2}\.\d{4} \d{2}:\d{2}) do (\d{2}\.\d{2}\.\d{4} \d{2}:\d{2})`)
	reUntilFurtherNotice := regexp.MustCompile(`od (\d{2}\.\d{2}\.\d{4} \d{2}:\d{2}) do odvolání`)

	// Process each item in the RSS feed
	for _, item := range rssFeed.Channel.Items {
		// Check if the item is from the current year
		parsedPubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return nil, s.Name(), fmt.Errorf("failed to parse publication date: %w", err)
		}
		if parsedPubDate.Year() != currentYear {
			continue
		}

		// Parse the start and end time from the title
		var startTime, endTime time.Time

		switch {
		case reMultiDay.MatchString(item.Title):
			matches := reMultiDay.FindStringSubmatch(item.Title)
			startTime, _ = time.Parse("02.01.2006 15:04", matches[1])
			endTime, _ = time.Parse("02.01.2006 15:04", matches[2])
		case reSingleDay.MatchString(item.Title):
			matches := reSingleDay.FindStringSubmatch(item.Title)
			startTime, _ = time.Parse("02.01.2006 15:04", matches[1])
			// End time only contains the time part, so we need to apply it to the same day as the start time
			endTime, _ = time.Parse("02.01.2006 15:04", fmt.Sprintf("%s %s", startTime.Format("02.01.2006"), matches[2]))
		case reUntilFurtherNotice.MatchString(item.Title):
			matches := reUntilFurtherNotice.FindStringSubmatch(item.Title)
			startTime, _ = time.Parse("02.01.2006 15:04", matches[1])
			// Set the end time to the end of the start day (23:59:59)
			endTime = time.Date(startTime.Year(), startTime.Month(), startTime.Day(), 23, 59, 59, 0, startTime.Location())
		default:
			// If no matching pattern, skip this item
			continue
		}

		// Strip HTML tags from the description
		description := stripHTML(item.Description)

		// Append the incident to the list
		incidents = append(incidents, api.Incident{
			Title:         item.Title,
			Description:   &description,
			StartTime:     startTime,
			EndTime:       &endTime,
			DeepLink:      fmt.Sprintf("%s %s", rssFeed.Channel.Link, item.Guid),
			Impact:        api.ImpactMaintenance, // Adjust this based on the content if necessary
			StatusPageUrl: statusPageUrl,
			Scraper:       s.Name(),
		})
	}

	if len(incidents) == 0 {
		return nil, s.Name(), fmt.Errorf("no incidents found")
	}

	return incidents, s.Name(), nil
}

// stripHTML strips HTML tags from a string
func stripHTML(input string) string {
	// Simplistic approach to remove HTML tags
	return strings.ReplaceAll(input, "<[^>]*>", "")
}
