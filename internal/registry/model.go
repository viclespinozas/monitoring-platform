// Package registry implements the service registry: the platform's source of
// truth for which services exist and how they should be health-checked.
//
// Layering: HTTP handlers -> Service (rules, validation) -> Repository (SQL).
package registry

import (
	"context"
	"errors"
	"time"
)

// Service is a registered service as stored and returned by the API.
type Service struct {
	ID                   string    `json:"id"`
	Name                 string    `json:"name"`
	Team                 string    `json:"team"`
	Description          string    `json:"description"`
	Tags                 []string  `json:"tags"`
	HealthURL            string    `json:"health_url"`
	CheckIntervalSeconds int       `json:"check_interval_seconds"`
	TimeoutSeconds       int       `json:"timeout_seconds"`
	ExpectedStatus       int       `json:"expected_status"`
	CreatedAt            time.Time `json:"created_at"`
	UpdatedAt            time.Time `json:"updated_at"`
}

// Input is the writable part of a Service, used for both create and update
// (PUT replaces the whole resource). Zero values of the three numeric fields
// mean "use the default".
type Input struct {
	Name                 string   `json:"name"`
	Team                 string   `json:"team"`
	Description          string   `json:"description"`
	Tags                 []string `json:"tags"`
	HealthURL            string   `json:"health_url"`
	CheckIntervalSeconds int      `json:"check_interval_seconds"`
	TimeoutSeconds       int      `json:"timeout_seconds"`
	ExpectedStatus       int      `json:"expected_status"`
}

// ListFilter selects and pages the services returned by List.
type ListFilter struct {
	Team   string
	Tag    string
	Limit  int
	Offset int
}

var (
	// ErrNotFound means no live service has the requested id.
	ErrNotFound = errors.New("service not found")
	// ErrNameTaken means another live service already uses the name.
	ErrNameTaken = errors.New("service name already registered")
)

// Repository is the storage contract. Implementations never return
// soft-deleted services.
type Repository interface {
	Create(ctx context.Context, in Input) (Service, error)
	Get(ctx context.Context, id string) (Service, error)
	List(ctx context.Context, f ListFilter) (items []Service, total int, err error)
	Update(ctx context.Context, id string, in Input) (Service, error)
	Delete(ctx context.Context, id string) error
}
