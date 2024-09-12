package api

import (
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pkg/errors"
)

type Impact string

const (
	ImpactMinor       Impact = "minor"
	ImpactMajor       Impact = "major"
	ImpactCritical    Impact = "critical"
	ImpactMaintenance Impact = "maintenance"
	ImpactNone        Impact = "none"
)

type HttpMethod string

const (
	MethodPost HttpMethod = "POST"
	MethodGet  HttpMethod = "GET"
	MethodHead HttpMethod = "HEAD"
)

var ErrInvalidImpact = errors.New("invalid impact")

func ParseImpact(impact string) (Impact, error) {
	switch impact {
	case "minor":
		return ImpactMinor, nil
	case "major":
		return ImpactMajor, nil
	case "critical":
		return ImpactCritical, nil
	case "maintenance":
		return ImpactMaintenance, nil
	case "none":
		return ImpactNone, nil
	default:
		return "", ErrInvalidImpact
	}
}

type IncidentEventArray []IncidentEvent

func (sla *IncidentEventArray) Scan(src interface{}) error {
	return json.Unmarshal(src.([]byte), &sla)
}

func (sla IncidentEventArray) Value() (driver.Value, error) {
	val, err := json.Marshal(sla)
	return string(val), err
}

type IncidentEvent struct {
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Time        time.Time `json:"time"`
}

func NewIncidentEvent(title string, description string, time time.Time) IncidentEvent {
	return IncidentEvent{
		Title:       title,
		Description: description,
		Time:        time,
	}
}

type Incident struct {
	Title                   string             `json:"title"`
	Components              []string           `gorm:"column:components;type:jsonb" json:"components"`
	Events                  IncidentEventArray `gorm:"column:events;type:jsonb" json:"events"`
	StartTime               time.Time          `gorm:"column:start_time;secondarykey" json:"startTime"`
	EndTime                 *time.Time         `gorm:"column:end_time;secondarykey" json:"endTime"`
	Description             *string            `gorm:"column:description" json:"description"`
	DeepLink                string             `gorm:"column:deep_link;primarykey" json:"deepLink"`
	Impact                  Impact             `gorm:"column:impact;secondarykey" json:"impact"`
	StatusPageUrl           string             `gorm:"column:status_page_url;secondarykey" json:"statusPageUrl"`
	NotificationJobsStarted bool               `gorm:"column:notification_jobs_started;secondarykey" json:"notificationJobsStarted"`
	Scraper                 string             `gorm:"column:scraper" json:"scraper"`
}

func NewIncident(title string, components []string, events []IncidentEvent, startTime time.Time, endTime *time.Time, description *string, deepLink string, impact Impact, statusPageUrl string, scraper string) Incident {
	return Incident{
		Title:         title,
		Components:    components,
		Events:        events,
		StartTime:     startTime,
		EndTime:       endTime,
		Description:   description,
		DeepLink:      deepLink,
		Impact:        impact,
		StatusPageUrl: statusPageUrl,
		Scraper:       scraper,
	}
}

type JSONMap map[string]string
type JSONStruct map[string]interface{}

// Scan implementuje deserializaci z JSONB do mapy
func (m *JSONMap) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type *map[string]string", value)
	}

	return json.Unmarshal(bytes, m)
}

// Value implementuje serializaci mapy do JSONB
func (m JSONMap) Value() (driver.Value, error) {
	return json.Marshal(m)
}

func (m *JSONStruct) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return fmt.Errorf("unsupported Scan, storing driver.Value type %T into type *map[string]string", value)
	}

	return json.Unmarshal(bytes, m)
}

// Value implementuje serializaci mapy do JSONB
func (m JSONStruct) Value() (driver.Value, error) {
	return json.Marshal(m)
}

type StatusPage struct {
	Name string `gorm:"secondarykey" json:"name"`
	URL  string `gorm:"primarykey" json:"url"`
	// Used to determine if we should run a scrape for this status page
	LastHistoricallyScraped time.Time `json:"lastHistoricallyScraped"`
	LastCurrentlyScraped    time.Time `json:"lastCurrentlyScraped"`
	// IsIndexed is used to determine if the status page has ever been indexed in the search engine successfully
	IsIndexed        bool       `json:"isIndexed"`
	PreferredScraper string     `json:"preferredScraper"`
	Headers          JSONMap    `gorm:"type:jsonb" json:"headers"`
	RequestPayload   JSONStruct `gorm:"type:jsonb" json:"payload"`
	Method           HttpMethod `json:"httpMethod"`
	ValidationRules  JSONStruct `gorm:"type:jsonb" json:"rules"`
}

func NewStatusPage(name string, url string) StatusPage {
	return StatusPage{
		Name:                    name,
		URL:                     url,
		LastHistoricallyScraped: time.Time{},
		LastCurrentlyScraped:    time.Time{},
	}
}
