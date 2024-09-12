package rest

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"regexp"
	"time"

	"github.com/metoro-io/statusphere/common/api"
	"github.com/metoro-io/statusphere/scraper/internal/scraper/providers"
	"go.uber.org/zap"
)

type RestProvider struct {
	logger     *zap.Logger
	httpClient *http.Client
}

func (s *RestProvider) Name() string {
	return string(providers.ProviderRest)
}

func NewRestProvider(logger *zap.Logger, httpClient *http.Client) *RestProvider {
	return &RestProvider{
		logger:     logger,
		httpClient: httpClient,
	}
}

func (s *RestProvider) ScrapeStatusPageCurrent(ctx context.Context, page api.StatusPage) ([]api.Incident, string, error) {
	resp, err := s.DoRequest(ctx, page)

	if err != nil {
		return nil, s.Name(), err
	}

	body, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		errorMessage := fmt.Sprintf("Non-200 response: %d, Body: %s", resp.StatusCode, string(body))
		s.logger.Error(errorMessage)
		incident := s.createIncident("Ivalid response code", errorMessage, "status_code", page.URL, err)
		return []api.Incident{incident}, s.Name(), fmt.Errorf("Non-200 response: %d, Body: %s", resp.StatusCode, string(body))
	}

	if page.Method == api.MethodHead {
		// u HEAD metody neresime obsah response, povazujeme za validni
		return []api.Incident{}, s.Name(), nil
	}
	var jsonResponse map[string]interface{}

	// Deserializace (unmarshalování) JSON odpovědi do mapy
	err = json.Unmarshal(body, &jsonResponse)
	if err != nil {
		errorMessage := fmt.Sprintf("Error unmarshaling JSON, Body: %s, Error: %w", string(body), err)
		s.logger.Error(errorMessage)
		incident := s.createIncident("Error unmarshaling JSON", errorMessage, "unmarshaling", page.URL, err)
		return []api.Incident{incident}, s.Name(), fmt.Errorf("Non-200 response: %d, Body: %s", resp.StatusCode, string(body))
	}

	if page.ValidationRules != nil {
		err = s.ValidateJSONResponse(jsonResponse, page.ValidationRules)
		if err != nil {
			errorMessage := fmt.Sprintf("Invalid JSON response, Body: %s, Error: %w", string(body), err)
			s.logger.Error(errorMessage)
			incident := s.createIncident("Invalid JSON response", errorMessage, "unmarshaling", page.URL, err)
			return []api.Incident{incident}, s.Name(), fmt.Errorf("Non-200 response: %d, Body: %s", resp.StatusCode, string(body))
		}
	}

	return []api.Incident{}, s.Name(), nil
}

func (s *RestProvider) DoRequest(ctx context.Context, page api.StatusPage) (*http.Response, error) {

	var jsonData []byte

	// request payload
	if page.RequestPayload != nil {
		var err error
		jsonData, err = json.Marshal(page.RequestPayload)
		if err != nil {
			return nil, err
		}
	}

	var method = api.MethodHead
	if page.Method != "" {
		method = page.Method
	}

	req, err := http.NewRequestWithContext(ctx, string(method), page.URL, bytes.NewBuffer(jsonData))
	if err != nil {
		return nil, err
	}

	// set headers, replace environment variable placeholders in headers
	for key, value := range page.Headers {
		req.Header.Set(key, s.replaceEnvVariables(value))
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return resp, err
	}

	return resp, nil
}

// replaceEnvVariables nahrazuje {env.<ENV_VAR>} v hodnotách za skutečné hodnoty environment proměnných
func (s *RestProvider) replaceEnvVariables(value string) string {
	// Regularni vyraz pro rozpoznani {env.VARIABLE_NAME}
	re := regexp.MustCompile(`{env\.([A-Z0-9_]+)}`)

	// Vyhledání všech výskytů a jejich nahrazení hodnotami z prostředí
	return re.ReplaceAllStringFunc(value, func(match string) string {
		// Extrahujeme název proměnné (např. SMARTFORM_AUTH_HEADER)
		envVar := re.FindStringSubmatch(match)
		if len(envVar) > 1 {
			// Získání hodnoty z environment proměnné
			return os.Getenv(envVar[1])
		}
		return match // Vrací původní hodnotu, pokud nic nenajde
	})
}

// ValidateJSONResponse ověřuje, zda odpověď z API odpovídá očekávaným validačním pravidlům.
func (s *RestProvider) ValidateJSONResponse(response, rules map[string]interface{}) error {
	for key, expectedValue := range rules {
		actualValue, exists := response[key]
		if !exists {
			return fmt.Errorf("missing key: %s", key)
		}

		switch expected := expectedValue.(type) {
		case map[string]interface{}:
			// Rekurzivní validace pro vnořené mapy
			if actualMap, ok := actualValue.(map[string]interface{}); ok {
				if err := s.ValidateJSONResponse(actualMap, expected); err != nil {
					return err
				}
			} else {
				return fmt.Errorf("invalid type for key: %s, expected map[string]interface{}", key)
			}
		default:
			// Porovnání hodnot
			if actualValue != expected {
				return fmt.Errorf("invalid value for key: %s, expected %v, got %v", key, expected, actualValue)
			}
		}
	}
	return nil
}

func (s *RestProvider) ScrapeStatusPageHistorical(ctx context.Context, url string) ([]api.Incident, string, error) {
	// Neimplementováno
	return []api.Incident{}, s.Name(), nil
}

func (s *RestProvider) createIncident(title string, desc string, code string, url string, err error) api.Incident {
	description := fmt.Sprintf("%s: %v", desc, err)
	incident := api.Incident{
		Title:         title,
		Description:   &description,
		StartTime:     time.Now(),
		StatusPageUrl: url,
		DeepLink:      fmt.Sprintf("%s/%s", url, code),
		Impact:        api.ImpactCritical,
		Scraper:       s.Name(),
	}
	return incident
}
