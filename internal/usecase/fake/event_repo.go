package fake

import (
	"context"
	"sort"
	"sync"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// MemoryEventRepository — in-memory EventRepository.
// Название уникально в рамках одного владельца
// (как UNIQUE(events.owner_id, events.name)); нарушение → domain.ErrConflict.
type MemoryEventRepository struct {
	mu     sync.Mutex
	nextID int64
	list   []*domain.Event
	byID   map[int64]*domain.Event
	byUUID map[string]*domain.Event
}

func NewEventRepository() *MemoryEventRepository {
	return &MemoryEventRepository{
		byID:   make(map[int64]*domain.Event),
		byUUID: make(map[string]*domain.Event),
	}
}

// GetByOwnerID возвращает события владельца, отсортированные по ID.
func (r *MemoryEventRepository) GetByOwnerID(_ context.Context, ownerID int64) ([]domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Event, 0)
	for _, e := range r.list {
		if e.OwnerID == ownerID {
			out = append(out, *e)
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].ID < out[j].ID })
	return out, nil
}

func (r *MemoryEventRepository) GetByUUID(_ context.Context, uuid string) (*domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.byUUID[uuid]; ok {
		return e, nil
	}
	return nil, domain.ErrNotFound
}

func (r *MemoryEventRepository) GetByID(_ context.Context, id int64) (*domain.Event, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if e, ok := r.byID[id]; ok {
		return e, nil
	}
	return nil, domain.ErrNotFound
}

func (r *MemoryEventRepository) Create(_ context.Context, e *domain.Event) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for _, existing := range r.list {
		if existing.OwnerID == e.OwnerID && existing.Name == e.Name {
			return 0, domain.ErrConflict
		}
	}
	r.nextID++
	e.ID = r.nextID
	r.list = append(r.list, e)
	r.byID[e.ID] = e
	r.byUUID[e.UUID] = e
	return e.ID, nil
}

func (r *MemoryEventRepository) Update(_ context.Context, e *domain.Event) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.byID[e.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if e.OwnerID == existing.OwnerID && e.Name != existing.Name {
		for _, other := range r.list {
			if other.ID != e.ID && other.OwnerID == e.OwnerID && other.Name == e.Name {
				return domain.ErrConflict
			}
		}
	}
	if e.UUID != existing.UUID {
		delete(r.byUUID, existing.UUID)
		r.byUUID[e.UUID] = e
	}
	*existing = *e
	return nil
}
