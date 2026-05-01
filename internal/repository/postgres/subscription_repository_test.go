package postgres

import (
	"context"
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/google/uuid"

	domain "github.com/theun1c/effective-mobile-test-task/internal/domain/subscription"
	"github.com/theun1c/effective-mobile-test-task/internal/types/yearmonth"

	_ "github.com/jackc/pgx/v5/stdlib"
)

func TestSubscriptionRepositoryCreateAndGetByID(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	input := domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Yandex Plus",
		Price:       400,
		UserID:      uuid.New(),
		StartDate:   mustYearMonth(t, "07-2025"),
		EndDate:     yearMonthPointer(t, "12-2025"),
	}

	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if created.CreatedAt.IsZero() {
		t.Fatal("Create() returned zero CreatedAt")
	}

	if created.UpdatedAt.IsZero() {
		t.Fatal("Create() returned zero UpdatedAt")
	}

	got, err := repo.GetByID(ctx, input.ID)
	if err != nil {
		t.Fatalf("GetByID() error = %v", err)
	}

	assertSubscriptionsEqual(t, created, got)
}

func TestSubscriptionRepositoryGetByIDReturnsNotFound(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	repo := NewSubscriptionRepository(db)

	_, err := repo.GetByID(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByID() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestSubscriptionRepositoryListReturnsSubscriptionsOrderedByCreatedAtDesc(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	older, err := repo.Create(ctx, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Spotify",
		Price:       300,
		UserID:      uuid.New(),
		StartDate:   mustYearMonth(t, "01-2025"),
	})
	if err != nil {
		t.Fatalf("Create() older error = %v", err)
	}

	newer, err := repo.Create(ctx, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Netflix",
		Price:       900,
		UserID:      uuid.New(),
		StartDate:   mustYearMonth(t, "02-2025"),
		EndDate:     yearMonthPointer(t, "08-2025"),
	})
	if err != nil {
		t.Fatalf("Create() newer error = %v", err)
	}

	olderCreatedAt := time.Date(2025, time.January, 1, 10, 0, 0, 0, time.UTC)
	newerCreatedAt := time.Date(2025, time.February, 1, 10, 0, 0, 0, time.UTC)

	setSubscriptionTimestamps(t, db, older.ID, olderCreatedAt, olderCreatedAt)
	setSubscriptionTimestamps(t, db, newer.ID, newerCreatedAt, newerCreatedAt)

	list, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List() error = %v", err)
	}

	if len(list) != 2 {
		t.Fatalf("List() len = %d, want 2", len(list))
	}

	if list[0].ID != newer.ID || list[1].ID != older.ID {
		t.Fatalf("List() order = [%s, %s], want [%s, %s]", list[0].ID, list[1].ID, newer.ID, older.ID)
	}
}

func TestSubscriptionRepositoryUpdateReplacesSubscriptionFields(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Start Service",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   mustYearMonth(t, "03-2025"),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	updateInput := domain.Subscription{
		ID:          created.ID,
		ServiceName: "Updated Service",
		Price:       650,
		UserID:      uuid.New(),
		StartDate:   mustYearMonth(t, "04-2025"),
		EndDate:     yearMonthPointer(t, "11-2025"),
	}

	updated, err := repo.Update(ctx, updateInput)
	if err != nil {
		t.Fatalf("Update() error = %v", err)
	}

	if !updated.CreatedAt.Equal(created.CreatedAt) {
		t.Fatalf("Update() CreatedAt = %s, want %s", updated.CreatedAt, created.CreatedAt)
	}

	if !updated.UpdatedAt.After(created.UpdatedAt) && !updated.UpdatedAt.Equal(created.UpdatedAt) {
		t.Fatalf("Update() UpdatedAt = %s, want >= %s", updated.UpdatedAt, created.UpdatedAt)
	}

	if updated.ServiceName != updateInput.ServiceName {
		t.Fatalf("Update() ServiceName = %q, want %q", updated.ServiceName, updateInput.ServiceName)
	}

	if updated.Price != updateInput.Price {
		t.Fatalf("Update() Price = %d, want %d", updated.Price, updateInput.Price)
	}

	if updated.UserID != updateInput.UserID {
		t.Fatalf("Update() UserID = %s, want %s", updated.UserID, updateInput.UserID)
	}

	if updated.StartDate.String() != updateInput.StartDate.String() {
		t.Fatalf("Update() StartDate = %s, want %s", updated.StartDate, updateInput.StartDate)
	}

	if updated.EndDate == nil || updated.EndDate.String() != updateInput.EndDate.String() {
		t.Fatalf("Update() EndDate = %v, want %v", updated.EndDate, updateInput.EndDate)
	}
}

func TestSubscriptionRepositoryUpdateReturnsNotFound(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	repo := NewSubscriptionRepository(db)

	_, err := repo.Update(context.Background(), domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Missing",
		Price:       100,
		UserID:      uuid.New(),
		StartDate:   mustYearMonth(t, "03-2025"),
	})
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Update() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestSubscriptionRepositoryDeleteRemovesSubscription(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	repo := NewSubscriptionRepository(db)
	ctx := context.Background()

	created, err := repo.Create(ctx, domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Delete Me",
		Price:       500,
		UserID:      uuid.New(),
		StartDate:   mustYearMonth(t, "05-2025"),
	})
	if err != nil {
		t.Fatalf("Create() error = %v", err)
	}

	if err := repo.Delete(ctx, created.ID); err != nil {
		t.Fatalf("Delete() error = %v", err)
	}

	_, err = repo.GetByID(ctx, created.ID)
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("GetByID() after Delete error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestSubscriptionRepositoryDeleteReturnsNotFound(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	repo := NewSubscriptionRepository(db)

	err := repo.Delete(context.Background(), uuid.New())
	if !errors.Is(err, domain.ErrNotFound) {
		t.Fatalf("Delete() error = %v, want %v", err, domain.ErrNotFound)
	}
}

func TestSubscriptionRepositoryCreateMapsCheckViolationToValidationError(t *testing.T) {
	db := openTestDB(t)
	resetSubscriptionsSchema(t, db)

	repo := NewSubscriptionRepository(db)

	_, err := repo.Create(context.Background(), domain.Subscription{
		ID:          uuid.New(),
		ServiceName: "Invalid Price",
		Price:       0,
		UserID:      uuid.New(),
		StartDate:   mustYearMonth(t, "06-2025"),
	})
	if !errors.Is(err, domain.ErrValidation) {
		t.Fatalf("Create() error = %v, want %v", err, domain.ErrValidation)
	}
}

func openTestDB(t *testing.T) *sql.DB {
	t.Helper()

	dsn := os.Getenv("TEST_POSTGRES_DSN")
	if dsn == "" {
		dsn = "postgres://subscriptions:subscriptions@127.0.0.1:5432/subscriptions?sslmode=disable"
	}

	db, err := sql.Open("pgx", dsn)
	if err != nil {
		t.Fatalf("sql.Open() error = %v", err)
	}

	t.Cleanup(func() {
		_ = db.Close()
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		t.Fatalf("db.PingContext() error = %v", err)
	}

	return db
}

func resetSubscriptionsSchema(t *testing.T, db *sql.DB) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if _, err := db.ExecContext(ctx, `DROP TABLE IF EXISTS subscriptions CASCADE`); err != nil {
		t.Fatalf("drop subscriptions table: %v", err)
	}

	migrationPath := filepath.Join(repositoryRoot(t), "migrations", "000001_create_subscriptions.up.sql")
	migration, err := os.ReadFile(migrationPath)
	if err != nil {
		t.Fatalf("read migration %s: %v", migrationPath, err)
	}

	if _, err := db.ExecContext(ctx, string(migration)); err != nil {
		t.Fatalf("apply migration: %v", err)
	}
}

func repositoryRoot(t *testing.T) string {
	t.Helper()

	_, file, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("runtime.Caller() failed")
	}

	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", ".."))
}

func setSubscriptionTimestamps(t *testing.T, db *sql.DB, id uuid.UUID, createdAt, updatedAt time.Time) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	_, err := db.ExecContext(
		ctx,
		`UPDATE subscriptions SET created_at = $2, updated_at = $3 WHERE id = $1`,
		id,
		createdAt,
		updatedAt,
	)
	if err != nil {
		t.Fatalf("set timestamps for %s: %v", id, err)
	}
}

func assertSubscriptionsEqual(t *testing.T, want, got domain.Subscription) {
	t.Helper()

	if want.ID != got.ID {
		t.Fatalf("ID = %s, want %s", got.ID, want.ID)
	}

	if want.ServiceName != got.ServiceName {
		t.Fatalf("ServiceName = %q, want %q", got.ServiceName, want.ServiceName)
	}

	if want.Price != got.Price {
		t.Fatalf("Price = %d, want %d", got.Price, want.Price)
	}

	if want.UserID != got.UserID {
		t.Fatalf("UserID = %s, want %s", got.UserID, want.UserID)
	}

	if want.StartDate.String() != got.StartDate.String() {
		t.Fatalf("StartDate = %s, want %s", got.StartDate, want.StartDate)
	}

	switch {
	case want.EndDate == nil && got.EndDate != nil:
		t.Fatalf("EndDate = %v, want nil", got.EndDate)
	case want.EndDate != nil && got.EndDate == nil:
		t.Fatalf("EndDate = nil, want %s", want.EndDate)
	case want.EndDate != nil && got.EndDate != nil && want.EndDate.String() != got.EndDate.String():
		t.Fatalf("EndDate = %s, want %s", got.EndDate, want.EndDate)
	}

	if !want.CreatedAt.Equal(got.CreatedAt) {
		t.Fatalf("CreatedAt = %s, want %s", got.CreatedAt, want.CreatedAt)
	}

	if !want.UpdatedAt.Equal(got.UpdatedAt) {
		t.Fatalf("UpdatedAt = %s, want %s", got.UpdatedAt, want.UpdatedAt)
	}
}

func mustYearMonth(t *testing.T, value string) yearmonth.YearMonth {
	t.Helper()

	parsed, err := yearmonth.Parse(value)
	if err != nil {
		t.Fatalf("yearmonth.Parse(%q) error = %v", value, err)
	}

	return parsed
}

func yearMonthPointer(t *testing.T, value string) *yearmonth.YearMonth {
	t.Helper()

	parsed := mustYearMonth(t, value)
	return &parsed
}
