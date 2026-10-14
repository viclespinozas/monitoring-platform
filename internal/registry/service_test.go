package registry

import (
	"errors"
	"strings"
	"testing"
)

func validInput() Input {
	return Input{Name: "orders", Team: "core", HealthURL: "http://orders:8080/healthz"}
}

func TestNormalizeAppliesDefaults(t *testing.T) {
	got, err := normalize(validInput())
	if err != nil {
		t.Fatal(err)
	}
	if got.CheckIntervalSeconds != 30 || got.TimeoutSeconds != 5 || got.ExpectedStatus != 200 {
		t.Fatalf("defaults not applied: %+v", got)
	}
}

func TestNormalizeTrimsAndDedupesTags(t *testing.T) {
	in := validInput()
	in.Name = "  orders "
	in.Tags = []string{"prod", " prod ", "tier:1"}
	got, err := normalize(in)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "orders" {
		t.Errorf("name not trimmed: %q", got.Name)
	}
	if len(got.Tags) != 2 || got.Tags[0] != "prod" || got.Tags[1] != "tier:1" {
		t.Errorf("tags = %v", got.Tags)
	}
}

func TestNormalizeRejects(t *testing.T) {
	tests := []struct {
		name      string
		mutate    func(*Input)
		wantField string
	}{
		{"empty name", func(i *Input) { i.Name = "" }, "name"},
		{"uppercase name", func(i *Input) { i.Name = "Orders" }, "name"},
		{"leading hyphen", func(i *Input) { i.Name = "-orders" }, "name"},
		{"name too long", func(i *Input) { i.Name = strings.Repeat("a", 101) }, "name"},
		{"empty team", func(i *Input) { i.Team = " " }, "team"},
		{"long description", func(i *Input) { i.Description = strings.Repeat("x", 501) }, "description"},
		{"relative url", func(i *Input) { i.HealthURL = "/healthz" }, "health_url"},
		{"ftp url", func(i *Input) { i.HealthURL = "ftp://host/x" }, "health_url"},
		{"interval too small", func(i *Input) { i.CheckIntervalSeconds = 2 }, "check_interval_seconds"},
		{"interval too large", func(i *Input) { i.CheckIntervalSeconds = 4000 }, "check_interval_seconds"},
		{"timeout not below interval", func(i *Input) { i.CheckIntervalSeconds = 10; i.TimeoutSeconds = 10 }, "timeout_seconds"},
		{"bad status", func(i *Input) { i.ExpectedStatus = 700 }, "expected_status"},
		{"bad tag", func(i *Input) { i.Tags = []string{"Has Space"} }, "tags[0]"},
		{"too many tags", func(i *Input) {
			for n := 0; n < 21; n++ {
				i.Tags = append(i.Tags, "t"+strings.Repeat("x", n))
			}
		}, "tags"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			in := validInput()
			tt.mutate(&in)
			_, err := normalize(in)
			var ve *ValidationError
			if !errors.As(err, &ve) {
				t.Fatalf("expected ValidationError, got %v", err)
			}
			for _, f := range ve.Fields {
				if f.Field == tt.wantField {
					return
				}
			}
			t.Fatalf("no error for field %q in %v", tt.wantField, ve.Fields)
		})
	}
}

func TestNormalizeCollectsAllErrors(t *testing.T) {
	_, err := normalize(Input{})
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("expected ValidationError, got %v", err)
	}
	if len(ve.Fields) < 3 {
		t.Fatalf("expected several field errors at once, got %v", ve.Fields)
	}
}

func TestListFilterNormalized(t *testing.T) {
	tests := []struct {
		in         ListFilter
		wantLimit  int
		wantOffset int
	}{
		{ListFilter{}, DefaultPageSize, 0},
		{ListFilter{Limit: 500}, MaxPageSize, 0},
		{ListFilter{Limit: 5, Offset: -3}, 5, 0},
		{ListFilter{Limit: 5, Offset: 40}, 5, 40},
	}
	for _, tt := range tests {
		got := tt.in.Normalized()
		if got.Limit != tt.wantLimit || got.Offset != tt.wantOffset {
			t.Errorf("Normalized(%+v) = %+v", tt.in, got)
		}
	}
}
