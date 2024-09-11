package instacover

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/metoro-io/statusphere/common/api"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type APIProvider struct {
	logger     *zap.Logger
	httpClient *http.Client
}

func NewAPIProvider(logger *zap.Logger, httpClient *http.Client) *APIProvider {
	return &APIProvider{
		logger:     logger,
		httpClient: httpClient,
	}
}

func (p *APIProvider) Name() string {
	return "InstacoverAPI"
}

// Structure to hold the JSON response
type ApiResponse struct {
	Status string `json:"status"`
}

// ScrapeStatusPageCurrent checks the status page using a REST API
func (s *APIProvider) ScrapeStatusPageCurrent(ctx context.Context, url string) ([]api.Incident, string, error) {
	// Perform the API call
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, s.Name(), errors.Wrap(err, "failed to create request")
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		s.logger.Error("API request failed", zap.String("url", url), zap.Error(err))
		var description = fmt.Sprintf("Failed to reach the service at %w", url)
		return []api.Incident{
			s.createIncident("API request failed", description, url, "server_error"),
		}, s.Name(), nil
	}
	defer resp.Body.Close()

	// If the status code is not 200, we create an incident
	if resp.StatusCode != http.StatusOK {
		s.logger.Error("Non-200 status code", zap.String("url", url), zap.Int("status_code", resp.StatusCode))
		var description = fmt.Sprintf("Service returned non-200 status code: %d", resp.StatusCode)
		return []api.Incident{
			s.createIncident("API Service Down", description, url, fmt.Sprintf("status_code_%w", resp.StatusCode)),
		}, s.Name(), nil
	}

	// Parse the JSON response
	var apiResp ApiResponse
	if err := json.NewDecoder(resp.Body).Decode(&apiResp); err != nil {
		s.logger.Error("Failed to parse JSON response", zap.String("url", url), zap.Error(err))
		return nil, s.Name(), errors.Wrap(err, "failed to parse API response")
	}

	// If status is not "OK", we consider it an incident
	status := strings.ToLower(apiResp.Status)
	if status != "ok" {
		s.logger.Error("Service returned non-OK status", zap.String("url", url), zap.String("status", apiResp.Status))
		var description = fmt.Sprintf("Service at %s returned status: %s", url, apiResp.Status)
		return []api.Incident{
			s.createIncident("Service Degraded", description, url, apiResp.Status),
		}, s.Name(), nil
	}

	// No incidents if everything is OK
	return []api.Incident{}, s.Name(), nil
}

func (p *APIProvider) createIncident(title string, description string, url string, linkCode string) api.Incident {
	return api.Incident{
		Title:         title,
		Description:   &description,
		StartTime:     time.Now(),
		StatusPageUrl: url,
		DeepLink:      fmt.Sprintf("%s/%s", url, linkCode),
		Impact:        api.ImpactCritical,
		Scraper:       p.Name(),
	}

}

// There is no historical page differentiation for API, we skip this
func (s *APIProvider) ScrapeStatusPageHistorical(ctx context.Context, url string) ([]api.Incident, string, error) {
	return []api.Incident{}, s.Name(), nil
}
