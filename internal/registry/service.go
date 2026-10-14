package registry

import (
	"context"
	"fmt"
	"net/url"
	"regexp"
	"strings"
)

const (
	defaultCheckInterval = 30
	defaultTimeout       = 5
	defaultStatus        = 200

	DefaultPageSize = 20
	MaxPageSize     = 100
)

var (
	nameRe = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]{0,98}[a-z0-9])?$`)
	tagRe  = regexp.MustCompile(`^[a-z0-9][a-z0-9:_-]{0,49}$`)
)

// FieldError describes one invalid field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// ValidationError carries every invalid field so clients can fix them in one go.
type ValidationError struct {
	Fields []FieldError
}

func (e *ValidationError) Error() string {
	parts := make([]string, len(e.Fields))
	for i, f := range e.Fields {
		parts[i] = f.Field + ": " + f.Message
	}
	return "validation failed: " + strings.Join(parts, "; ")
}

// Manager applies business rules on top of a Repository.
type Manager struct {
	repo Repository
}

// NewManager returns a Manager backed by repo.
func NewManager(repo Repository) *Manager {
	return &Manager{repo: repo}
}

func (m *Manager) Create(ctx context.Context, in Input) (Service, error) {
	clean, err := normalize(in)
	if err != nil {
		return Service{}, err
	}
	return m.repo.Create(ctx, clean)
}

func (m *Manager) Get(ctx context.Context, id string) (Service, error) {
	return m.repo.Get(ctx, id)
}

func (m *Manager) List(ctx context.Context, f ListFilter) ([]Service, int, error) {
	return m.repo.List(ctx, f.Normalized())
}

func (m *Manager) Update(ctx context.Context, id string, in Input) (Service, error) {
	clean, err := normalize(in)
	if err != nil {
		return Service{}, err
	}
	return m.repo.Update(ctx, id, clean)
}

func (m *Manager) Delete(ctx context.Context, id string) error {
	return m.repo.Delete(ctx, id)
}

// Normalized applies the default page size and clamps limit and offset.
func (f ListFilter) Normalized() ListFilter {
	if f.Limit <= 0 {
		f.Limit = DefaultPageSize
	}
	if f.Limit > MaxPageSize {
		f.Limit = MaxPageSize
	}
	if f.Offset < 0 {
		f.Offset = 0
	}
	return f
}

// normalize trims, applies defaults and validates. It collects all problems.
func normalize(in Input) (Input, error) {
	var errs []FieldError
	bad := func(field, msg string) { errs = append(errs, FieldError{field, msg}) }

	in.Name = strings.TrimSpace(in.Name)
	in.Team = strings.TrimSpace(in.Team)
	in.Description = strings.TrimSpace(in.Description)
	in.HealthURL = strings.TrimSpace(in.HealthURL)

	if !nameRe.MatchString(in.Name) {
		bad("name", "must be 1-100 characters: lowercase letters, digits and hyphens, not starting or ending with a hyphen")
	}
	if in.Team == "" || len(in.Team) > 100 {
		bad("team", "is required and must be at most 100 characters")
	}
	if len(in.Description) > 500 {
		bad("description", "must be at most 500 characters")
	}

	if u, err := url.Parse(in.HealthURL); err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" || len(in.HealthURL) > 2048 {
		bad("health_url", "must be an absolute http or https URL")
	}

	if in.CheckIntervalSeconds == 0 {
		in.CheckIntervalSeconds = defaultCheckInterval
	}
	if in.TimeoutSeconds == 0 {
		in.TimeoutSeconds = defaultTimeout
	}
	if in.ExpectedStatus == 0 {
		in.ExpectedStatus = defaultStatus
	}
	if in.CheckIntervalSeconds < 5 || in.CheckIntervalSeconds > 3600 {
		bad("check_interval_seconds", "must be between 5 and 3600")
	}
	if in.TimeoutSeconds < 1 || in.TimeoutSeconds > 300 {
		bad("timeout_seconds", "must be between 1 and 300")
	}
	if in.TimeoutSeconds >= in.CheckIntervalSeconds {
		bad("timeout_seconds", "must be smaller than check_interval_seconds")
	}
	if in.ExpectedStatus < 100 || in.ExpectedStatus > 599 {
		bad("expected_status", "must be a valid HTTP status code (100-599)")
	}

	tags, tagErrs := normalizeTags(in.Tags)
	in.Tags = tags
	errs = append(errs, tagErrs...)

	if len(errs) > 0 {
		return Input{}, &ValidationError{Fields: errs}
	}
	return in, nil
}

func normalizeTags(tags []string) ([]string, []FieldError) {
	if len(tags) > 20 {
		return nil, []FieldError{{"tags", "at most 20 tags are allowed"}}
	}
	out := make([]string, 0, len(tags))
	seen := make(map[string]bool, len(tags))
	var errs []FieldError
	for i, t := range tags {
		t = strings.TrimSpace(t)
		if !tagRe.MatchString(t) {
			errs = append(errs, FieldError{fmt.Sprintf("tags[%d]", i), "must be 1-50 characters: lowercase letters, digits, '-', '_' or ':'"})
			continue
		}
		if !seen[t] {
			seen[t] = true
			out = append(out, t)
		}
	}
	return out, errs
}
