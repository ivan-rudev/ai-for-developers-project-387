package sqlite

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/infrastructure/config"
)

func TestSeedSmoke(t *testing.T) {
	dir := t.TempDir()
	cfgPath := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(cfgPath, []byte(`
default:
  work_start: "09:00"
  work_end: "18:00"
  timezone: "Europe/Moscow"
  working_days: [mon, tue, wed, thu, fri]
  events:
    - name: "Короткая встреча"
      duration_minutes: 15
    - name: "Стандартная встреча"
      duration_minutes: 30
seed:
  owners:
    - uuid: "550e8400-e29b-41d4-a716-446655440000"
      name: "Bob Jones"
      email: "bob@example.com"
  events:
    - owner_uuid: "550e8400-e29b-41d4-a716-446655440000"
      name: "Короткая встреча"
      duration_minutes: 15
    - owner_uuid: "550e8400-e29b-41d4-a716-446655440000"
      name: "Стандартная встреча"
      duration_minutes: 30
`), 0o644); err != nil {
		t.Fatal(err)
	}

	cfg, err := config.Load(cfgPath)
	if err != nil {
		t.Fatal(err)
	}

	db, err := Open(filepath.Join(dir, "data.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer db.Close()

	ctx := context.Background()
	if err := db.Seed(ctx, cfg); err != nil {
		t.Fatalf("seed: %v", err)
	}

	owners := NewOwnerRepository(db)
	owner, err := owners.GetByUUID(ctx, "550e8400-e29b-41d4-a716-446655440000")
	if err != nil {
		t.Fatalf("get owner: %v", err)
	}
	if owner.Name != "Bob Jones" || owner.WorkStart != "09:00" || owner.Timezone != "Europe/Moscow" {
		t.Fatalf("unexpected owner: %+v", owner)
	}
	if !owner.WorkingDays["mon"] || owner.WorkingDays["sun"] {
		t.Fatalf("unexpected working days: %+v", owner.WorkingDays)
	}

	events := NewEventRepository(db)
	list, err := events.GetByOwnerID(ctx, owner.ID)
	if err != nil {
		t.Fatalf("list events: %v", err)
	}
	if len(list) != 2 {
		t.Fatalf("expected 2 events, got %d", len(list))
	}

	if err := db.Seed(ctx, cfg); err != nil {
		t.Fatalf("seed second time: %v", err)
	}
	list2, err := events.GetByOwnerID(ctx, owner.ID)
	if err != nil {
		t.Fatalf("list events after re-seed: %v", err)
	}
	if len(list2) != 2 {
		t.Fatalf("re-seed duplicated events: got %d, want 2", len(list2))
	}
}
