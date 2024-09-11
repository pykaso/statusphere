package server

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/metoro-io/statusphere/common/api"
	"github.com/patrickmn/go-cache"
	"github.com/pkg/errors"
	"go.uber.org/zap"
)

type Status string

const (
	StatusUp       Status = "UP"
	StatusDegraded Status = "DEGRADED"
	StatusUnknown  Status = "UNKNOWN"
)

type CurrentStatusResponse struct {
	Status    Status `json:"status"`
	IsIndexed bool   `json:"isIndexed"`
}

// currentStatus is a handler for the /current-status endpoint.
// It has a required query parameter of statusPageUrl
// It returns the current status of the status page.
// If the status page is not known to statusphere, it returns a 404.
// If the status page is known to statusphere and it is indexed, it returns UP or DEGRADED depending on the current incidents.
// If the status page is known to statusphere and it is not indexed, it returns UNKNOWN.
func (s *Server) currentStatus(context *gin.Context) {
	ctx := context.Request.Context()
	statusPageUrl := context.Query("statusPageUrl")
	if statusPageUrl == "" {
		context.JSON(http.StatusBadRequest, gin.H{"error": "statusPageUrl is required"})
		return
	}

	statusPageInterface, found := s.statusPageCache.Get(statusPageUrl)
	if !found {
		context.JSON(http.StatusNotFound, gin.H{"error": "status page not known to statusphere"})
		return
	}

	statusPageInterfaceCasted, ok := statusPageInterface.(api.StatusPage)
	if !ok {
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to cast status page to api.StatusPage"})
		return
	}

	if !statusPageInterfaceCasted.IsIndexed {
		context.JSON(http.StatusOK, CurrentStatusResponse{Status: StatusUnknown, IsIndexed: false})
		return
	}

	// Attempt to get the incidents from the cache
	incidents, found, err := s.getCurrentIncidentsFromCache(ctx, statusPageUrl)
	if err != nil {
		s.logger.Error("failed to get incidents from cache", zap.Error(err))
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get incidents from cache"})
		return
	}
	if found {
		if len(incidents) > 0 {
			context.JSON(http.StatusOK, CurrentStatusResponse{Status: StatusDegraded, IsIndexed: true})
			return
		}
		context.JSON(http.StatusOK, CurrentStatusResponse{Status: StatusUp, IsIndexed: true})
		return
	}

	// Attempt to get the incidents from the database
	incidents, found, err = s.getCurrentIncidentsFromDatabase(ctx, statusPageUrl)
	if err != nil {
		s.logger.Error("failed to get incidents from database", zap.Error(err))
		context.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get incidents from database"})
		return
	}
	if !found {
		context.JSON(http.StatusNotFound, gin.H{"error": "status page not known to statusphere"})
		return
	}

	s.currentIncidentCache.Set(statusPageUrl, incidents, cache.DefaultExpiration)
	if len(incidents) > 0 {
		context.JSON(http.StatusOK, CurrentStatusResponse{Status: StatusDegraded, IsIndexed: true})
		return
	}

	context.JSON(http.StatusOK, CurrentStatusResponse{Status: StatusUp, IsIndexed: true})
}

// getCurrentIncidentsFromCache attempts to get the current incidents from the cache.
// If the incidents are found in the cache, it returns them.
// If the incidents are not found in the cache, it returns false for the second return value.

func (s *Server) getCurrentIncidentsFromCache(ctx context.Context, statusPageUrl string) ([]api.Incident, bool, error) {
	incidents, found := s.currentIncidentCache.Get(statusPageUrl)
	if !found {
		return nil, false, nil
	}

	incidentsCasted, ok := incidents.([]api.Incident)
	if !ok {
		return nil, false, errors.New("failed to cast incidents to []api.Incident")
	}

	return incidentsCasted, true, nil
}

// getCurrentIncidentsFromDatabase attempts to get the current incidents from the database.
// If the incidents are found in the database, it returns them.
// If the incidents are not found in the database, it returns false for the second return value.
func (s *Server) getCurrentIncidentsFromDatabase(ctx context.Context, statusPageUrl string) ([]api.Incident, bool, error) {
	incidents, err := s.dbClient.GetCurrentIncidents(ctx, statusPageUrl)
	if err != nil {
		return nil, false, err
	}

	if len(incidents) == 0 {
		// See if the status page exists
		statusPage, err := s.dbClient.GetStatusPage(ctx, statusPageUrl)
		if err != nil {
			return nil, false, err
		}
		if statusPage == nil {
			// The status page does not exist
			return nil, false, nil
		}
	}

	return incidents, true, nil
}
