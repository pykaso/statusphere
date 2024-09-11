package db

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/kelseyhightower/envconfig"
	"github.com/metoro-io/statusphere/common/api"
	"github.com/metoro-io/statusphere/common/status_pages"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"gorm.io/gorm/logger"
)

const schemaName = "statusphere"

type Config struct {
	Host     string `envconfig:"POSTGRES_HOST"`
	Port     string `envconfig:"POSTGRES_PORT"`
	User     string `envconfig:"POSTGRES_USER"`
	Password string `envconfig:"POSTGRES_PASSWORD"`
	Database string `envconfig:"POSTGRES_DATABASE"`
}

func getConfigFromEnvironment() (Config, error) {
	var config Config
	err := envconfig.Process("STATUSPHERE", &config)
	return config, err
}

type DbClient struct {
	PgxPool *pgxpool.Pool
	db      *gorm.DB
	logger  *zap.Logger
}

func NewDbClientFromEnvironment(lg *zap.Logger) (*DbClient, error) {
	config, err := getConfigFromEnvironment()
	if err != nil {
		return nil, err
	}

	// Check to see if the database exists in postgres
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Database)
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to postgres")
	}
	// Create the database if it does not exist
	err = db.Exec(fmt.Sprintf("CREATE DATABASE %s", config.Database)).Error
	wasCreatedSuccessfully := false
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			if pgErr.Code == "42P04" {
				// This is the code for database already exists
				// We can ignore this error
				wasCreatedSuccessfully = true
			}
		}
		if !wasCreatedSuccessfully {
			return nil, errors.Wrap(err, "failed to create postgres database")
		}
	}

	// Connect to the database
	dsn = fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		config.Host, config.Port, config.User, config.Password, config.Database)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,   // Slow SQL threshold
			LogLevel:                  logger.Silent, // Log level
			IgnoreRecordNotFoundError: true,          // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,          // Don't include params in the SQL log
			Colorful:                  false,         // Disable color
		},
	)
	db, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return nil, errors.Wrap(err, "failed to connect to postgres")
	}

	pgxPool, err := pgxpool.New(context.Background(), dsn)
	if err != nil {
		return nil, errors.Wrap(err, "failed to create pgx pool")
	}

	return &DbClient{db: db, logger: lg, PgxPool: pgxPool}, nil
}

const statusPageTableName = "status_page"
const incidentsTableName = "incidents"

func (d *DbClient) AutoMigrate(ctx context.Context) error {
	d.logger.Info("DbClient.AutoMigrate()")
	// Create the schema if it does not exist
	d.db.Exec(fmt.Sprintf("CREATE SCHEMA IF NOT EXISTS %s", schemaName))

	// Create the statuspage table
	err := d.db.Table(fmt.Sprintf(fmt.Sprintf("%s.%s", schemaName, statusPageTableName))).AutoMigrate(&api.StatusPage{})
	if err != nil {
		return errors.Wrap(err, "failed to auto-migrate status_page table")
	}
	err = d.SeedStatusPages()
	if err != nil {
		return errors.Wrap(err, "failed to seed status pages")
	}

	// Create the incidents table
	err = d.db.Table(fmt.Sprintf(fmt.Sprintf("%s.%s", schemaName, incidentsTableName))).AutoMigrate(&api.Incident{})
	if err != nil {
		return errors.Wrap(err, "failed to auto-migrate incidents table")
	}

	return nil
}

func (d *DbClient) GetAllStatusPages(ctx context.Context) ([]api.StatusPage, error) {
	var statusPages []api.StatusPage
	result := d.db.Table(fmt.Sprintf(fmt.Sprintf("%s.%s", schemaName, statusPageTableName))).Find(&statusPages)
	if result.Error != nil {
		return nil, result.Error
	}
	return statusPages, nil
}

func (d *DbClient) GetStatusPage(ctx context.Context, url string) (*api.StatusPage, error) {
	var statusPage api.StatusPage
	result := d.db.Table(fmt.Sprintf(fmt.Sprintf("%s.%s", schemaName, statusPageTableName))).Where("url = ?", url).First(&statusPage)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, result.Error
	}
	return &statusPage, nil
}

func (d *DbClient) UpdateStatusPage(ctx context.Context, statusPage api.StatusPage) error {
	result := d.db.Table(fmt.Sprintf(fmt.Sprintf("%s.%s", schemaName, statusPageTableName))).Where("url = ?", statusPage.URL).Updates(&statusPage)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *DbClient) InsertStatusPage(ctx context.Context, statusPage api.StatusPage) error {
	result := d.db.Table(fmt.Sprintf(fmt.Sprintf("%s.%s", schemaName, statusPageTableName))).Create(&statusPage)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

func (d *DbClient) GetIncidents(ctx context.Context, statusPageUrl string) ([]api.Incident, error) {
	var incidents []api.Incident
	result := d.db.Table(fmt.Sprintf(fmt.Sprintf("%s.%s", schemaName, incidentsTableName))).Where("status_page_url = ?", statusPageUrl).Find(&incidents)
	if result.Error != nil {
		return nil, result.Error
	}
	return incidents, nil
}

// Current incidents are incidents that have not ended and have a start time in the last two weeks
// The two week cutiff is not ideal but some incidents don't have a specified end time
func (d *DbClient) GetCurrentIncidentsOrig(ctx context.Context, statusPageUrl string) ([]api.Incident, error) {
	var incidents []api.Incident
	result := d.db.Table(fmt.Sprintf("%s.%s", schemaName, incidentsTableName)).Where("status_page_url = ? AND start_time > ? AND end_time = NULL", statusPageUrl, time.Now().Add(-14*24*time.Hour)).Find(&incidents)
	if result.Error != nil {
		return nil, result.Error
	}
	return incidents, nil
}

func (d *DbClient) GetCurrentIncidents(ctx context.Context, statusPageUrl string) ([]api.Incident, error) {
	var incidents []api.Incident

	// incidenty bez konce (end_time = NULL) nebo end_time v budoucnosti
	result := d.db.Table(fmt.Sprintf("%s.%s", schemaName, incidentsTableName)).
		Where("status_page_url = ? AND start_time <= ? AND (end_time IS NULL OR end_time > ?)",
			statusPageUrl, time.Now(), time.Now()).Find(&incidents)

	if result.Error != nil {
		return nil, result.Error
	}

	return incidents, nil
}

func (d *DbClient) GetIncidentsWithoutJobsStarted(ctx context.Context, limit int) ([]api.Incident, error) {
	var incidents []api.Incident
	result := d.db.Table(fmt.Sprintf("%s.%s", schemaName, incidentsTableName)).Where("notification_jobs_started is distinct from true limit ?", limit).Find(&incidents)
	if result.Error != nil {
		return nil, result.Error
	}
	return incidents, nil
}

func (d *DbClient) SetIncidentNotificationStartedToTrue(ctx context.Context, incidents []api.Incident) error {
	for i, _ := range incidents {
		incidents[i].NotificationJobsStarted = true
	}
	result := d.db.Table(fmt.Sprintf("%s.%s", schemaName, incidentsTableName)).Clauses(
		clause.OnConflict{
			Columns:   []clause.Column{{Name: "deep_link"}},                            // Primary key
			DoUpdates: clause.AssignmentColumns([]string{"notification_jobs_started"}), // Update the data column
		},
	).Create(&incidents)
	if result.Error != nil {
		return result.Error
	}
	return nil
}

// specificky scraper ktery neparsuje status page s incidenty ale overuje dostupnost adresy / API
func IsApiAvailabilityScraper(name string) bool {
	return name == "InstacoverAPI"
}

func generateRandomString(length int) (string, error) {
	bytes := make([]byte, length/2) // Protože každý bajt se bude reprezentovat dvěma hexadecimálními znaky
	_, err := rand.Read(bytes)
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *DbClient) ProcessAndCloseOngoingIncident(ctx context.Context, incidents []api.Incident, scraper string, url string) error {
	if len(incidents) > 0 {
		// pokud prijde incident, nedostupnost API stale trva
		return nil
	}

	// 1. Najdi probíhající incidenty (kde end_time == NULL) pro daný scraper
	var ongoingIncidents []api.Incident
	result := s.db.Table(fmt.Sprintf("%s.%s", schemaName, incidentsTableName)).
		Where("scraper = ? AND status_page_url = ? AND end_time IS NULL", scraper, url).Find(&ongoingIncidents)

	if result.Error != nil {
		return result.Error
	}

	now := time.Now()
	// Formátovaný čas bez prázdných řetězců
	formattedTime := now.Format("2006-01-02-15-04-05")

	// 2. Pokud nějaké probíhající incidenty existují uzavři je:
	// - nastav end_time na aktuální čas
	// - uprav deep_link ktery se pouziva jako identifikator incidentu
	if len(ongoingIncidents) > 0 {

		result = s.db.Table(fmt.Sprintf("%s.%s", schemaName, incidentsTableName)).
			Where("scraper = ? AND status_page_url = ? AND end_time IS NULL", scraper, url).
			Updates(map[string]interface{}{
				"end_time":  now,
				"deep_link": gorm.Expr("CONCAT(deep_link::text, ?::text)", "_"+formattedTime),
			})

		if result.Error != nil {
			return result.Error
		}
	}

	return nil
}

func (d *DbClient) CreateOrUpdateIncidents(ctx context.Context, incidents []api.Incident, scraper string, url string) error {
	if len(incidents) == 0 && IsApiAvailabilityScraper(scraper) {
		return d.ProcessAndCloseOngoingIncident(ctx, incidents, scraper, url)
	} else {
		// originalni implementace
		result := d.db.Table(fmt.Sprintf("%s.%s", schemaName, incidentsTableName)).Clauses(
			clause.OnConflict{
				Columns:   []clause.Column{{Name: "deep_link"}},                                                                                                                 // Primary key
				DoUpdates: clause.AssignmentColumns([]string{"title", "components", "events", "start_time", "end_time", "description", "impact", "status_page_url", "scraper"}), // Update the data column
			},
		).Create(&incidents)
		if result.Error != nil {
			return result.Error
		}
		return nil
	}
}

func (d *DbClient) SeedStatusPages() error {
	d.logger.Info("DbClient.SeedStatusPages()")

	for _, statusPage := range status_pages.StatusPages {
		if page, err := d.GetStatusPage(context.Background(), statusPage.URL); err != nil || page == nil {
			// Status page already exists
			err := d.InsertStatusPage(context.Background(), statusPage)
			if err != nil {
				return errors.Wrap(err, "failed to seed status pages")
			}
			d.logger.Info("Seeded status page: %s", zap.String("url", statusPage.URL))
		}
	}
	return nil
}

func (d *DbClient) DeleteStatusPage(background context.Context, url string) error {
	if url == "" {
		return errors.New("url cannot be empty")
	}
	result := d.db.Table(fmt.Sprintf("%s.%s", schemaName, statusPageTableName)).Where("url = ?", url).Delete(&api.StatusPage{})
	if result.Error != nil {
		return result.Error
	}
	return nil
}
