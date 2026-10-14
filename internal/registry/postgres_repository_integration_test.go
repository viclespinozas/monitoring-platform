//go:build integration

package registry_test

import (
	"context"
	"errors"
	"fmt"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/viclespinozas/monitoring-platform/internal/registry"
	"github.com/viclespinozas/monitoring-platform/internal/testutil/postgrestest"
)

var testDB *postgrestest.DB

func TestMain(m *testing.M) {
	ctx := context.Background()
	db, err := postgrestest.Start(ctx)
	if err != nil {
		fmt.Fprintln(os.Stderr, "cannot start postgres (is Docker running?):", err)
		os.Exit(1)
	}
	testDB = db
	code := m.Run()
	db.Stop(ctx)
	os.Exit(code)
}

func newRepo(t *testing.T) *registry.PostgresRepository {
	t.Helper()
	testDB.Reset(t)
	return registry.NewPostgresRepository(testDB.Pool)
}

func input(name string) registry.Input {
	return registry.Input{
		Name: name, Team: "core", Description: "d", Tags: []string{"prod"},
		HealthURL: "http://" + name + "/healthz", CheckIntervalSeconds: 30, TimeoutSeconds: 5, ExpectedStatus: 200,
	}
}

func TestCreateAndGetRoundTrip(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	in := input("orders")
	in.Tags = []string{"prod", "tier:1"}
	created, err := repo.Create(ctx, in)
	if err != nil {
		t.Fatal(err)
	}
	if created.ID == "" || created.CreatedAt.IsZero() || !created.CreatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("unexpected generated fields: %+v", created)
	}

	got, err := repo.Get(ctx, created.ID)
	if err != nil {
		t.Fatal(err)
	}
	if got.Name != "orders" || got.Team != "core" || got.HealthURL != in.HealthURL ||
		got.CheckIntervalSeconds != 30 || got.TimeoutSeconds != 5 || got.ExpectedStatus != 200 ||
		!slices.Equal(got.Tags, in.Tags) {
		t.Fatalf("round trip mismatch: %+v", got)
	}
}

func TestEmptyTagsComeBackAsEmptySlice(t *testing.T) {
	repo := newRepo(t)
	in := input("orders")
	in.Tags = []string{}
	created, err := repo.Create(context.Background(), in)
	if err != nil {
		t.Fatal(err)
	}
	if created.Tags == nil || len(created.Tags) != 0 {
		t.Fatalf("tags = %#v, want empty non-nil slice", created.Tags)
	}
}

func TestNameUniquenessAndSoftDelete(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	first, err := repo.Create(ctx, input("orders"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Create(ctx, input("orders")); !errors.Is(err, registry.ErrNameTaken) {
		t.Fatalf("duplicate live name: err = %v, want ErrNameTaken", err)
	}

	if err := repo.Delete(ctx, first.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Get(ctx, first.ID); !errors.Is(err, registry.ErrNotFound) {
		t.Fatalf("get after delete: err = %v, want ErrNotFound", err)
	}
	if err := repo.Delete(ctx, first.ID); !errors.Is(err, registry.ErrNotFound) {
		t.Fatalf("second delete: err = %v, want ErrNotFound", err)
	}
	if _, err := repo.Create(ctx, input("orders")); err != nil {
		t.Fatalf("name should be reusable after soft delete: %v", err)
	}
}

func TestNotFoundCases(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	tests := []struct {
		name string
		id   string
	}{
		{"unknown uuid", "00000000-0000-0000-0000-000000000000"},
		{"malformed id", "not-a-uuid"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if _, err := repo.Get(ctx, tt.id); !errors.Is(err, registry.ErrNotFound) {
				t.Errorf("Get: err = %v", err)
			}
			if _, err := repo.Update(ctx, tt.id, input("x")); !errors.Is(err, registry.ErrNotFound) {
				t.Errorf("Update: err = %v", err)
			}
			if err := repo.Delete(ctx, tt.id); !errors.Is(err, registry.ErrNotFound) {
				t.Errorf("Delete: err = %v", err)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	a, _ := repo.Create(ctx, input("alpha"))
	b, _ := repo.Create(ctx, input("beta"))

	time.Sleep(10 * time.Millisecond)
	changed := input("alpha")
	changed.Team = "payments"
	changed.CheckIntervalSeconds = 60
	updated, err := repo.Update(ctx, a.ID, changed)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Team != "payments" || updated.CheckIntervalSeconds != 60 {
		t.Fatalf("update not applied: %+v", updated)
	}
	if !updated.CreatedAt.Equal(a.CreatedAt) || !updated.UpdatedAt.After(a.UpdatedAt) {
		t.Fatalf("timestamps wrong: created %v->%v updated %v->%v", a.CreatedAt, updated.CreatedAt, a.UpdatedAt, updated.UpdatedAt)
	}

	rename := input("alpha")
	if _, err := repo.Update(ctx, b.ID, rename); !errors.Is(err, registry.ErrNameTaken) {
		t.Fatalf("rename onto existing name: err = %v, want ErrNameTaken", err)
	}

	if err := repo.Delete(ctx, a.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Update(ctx, a.ID, input("alpha")); !errors.Is(err, registry.ErrNotFound) {
		t.Fatalf("update deleted: err = %v, want ErrNotFound", err)
	}
}

func TestListFiltersPagingAndOrder(t *testing.T) {
	repo := newRepo(t)
	ctx := context.Background()

	seed := []struct{ name, team, tag string }{
		{"svc-a", "core", "prod"},
		{"svc-b", "core", "dev"},
		{"svc-c", "payments", "prod"},
		{"svc-d", "core", "prod"},
		{"svc-gone", "core", "prod"},
	}
	var goneID string
	for _, s := range seed {
		in := input(s.name)
		in.Team, in.Tags = s.team, []string{s.tag}
		created, err := repo.Create(ctx, in)
		if err != nil {
			t.Fatal(err)
		}
		if s.name == "svc-gone" {
			goneID = created.ID
		}
	}
	if err := repo.Delete(ctx, goneID); err != nil {
		t.Fatal(err)
	}

	names := func(items []registry.Service) []string {
		out := make([]string, len(items))
		for i, s := range items {
			out[i] = s.Name
		}
		return out
	}

	tests := []struct {
		name      string
		filter    registry.ListFilter
		wantNames []string
		wantTotal int
	}{
		{"all live, creation order", registry.ListFilter{Limit: 10}, []string{"svc-a", "svc-b", "svc-c", "svc-d"}, 4},
		{"by team", registry.ListFilter{Team: "core", Limit: 10}, []string{"svc-a", "svc-b", "svc-d"}, 3},
		{"by tag", registry.ListFilter{Tag: "prod", Limit: 10}, []string{"svc-a", "svc-c", "svc-d"}, 3},
		{"team and tag", registry.ListFilter{Team: "core", Tag: "prod", Limit: 10}, []string{"svc-a", "svc-d"}, 2},
		{"first page", registry.ListFilter{Limit: 2}, []string{"svc-a", "svc-b"}, 4},
		{"second page", registry.ListFilter{Limit: 2, Offset: 2}, []string{"svc-c", "svc-d"}, 4},
		{"past the end", registry.ListFilter{Limit: 2, Offset: 10}, []string{}, 4},
		{"no matches", registry.ListFilter{Team: "nobody", Limit: 10}, []string{}, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			items, total, err := repo.List(ctx, tt.filter)
			if err != nil {
				t.Fatal(err)
			}
			if total != tt.wantTotal || !slices.Equal(names(items), tt.wantNames) {
				t.Fatalf("got %v (total %d), want %v (total %d)", names(items), total, tt.wantNames, tt.wantTotal)
			}
		})
	}
}

func TestDatabaseConstraintsBackStopValidation(t *testing.T) {
	repo := newRepo(t)

	bad := input("orders")
	bad.CheckIntervalSeconds, bad.TimeoutSeconds = 10, 10
	_, err := repo.Create(context.Background(), bad)
	if err == nil {
		t.Fatal("expected the database to reject timeout >= interval")
	}
	if errors.Is(err, registry.ErrNameTaken) || errors.Is(err, registry.ErrNotFound) {
		t.Fatalf("constraint failure was mapped to a domain error: %v", err)
	}
}
