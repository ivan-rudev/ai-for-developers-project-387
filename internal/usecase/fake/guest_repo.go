package fake

import (
	"context"
	"sync"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// MemoryGuestRepository — in-memory GuestRepository.
// Email уникален (как UNIQUE(guests.email)); нарушение → domain.ErrConflict.
type MemoryGuestRepository struct {
	mu      sync.Mutex
	nextID  int64
	list    []*domain.Guest
	byID    map[int64]*domain.Guest
	byEmail map[string]*domain.Guest
}

func NewGuestRepository() *MemoryGuestRepository {
	return &MemoryGuestRepository{
		byID:    make(map[int64]*domain.Guest),
		byEmail: make(map[string]*domain.Guest),
	}
}

func (r *MemoryGuestRepository) GetByID(_ context.Context, id int64) (*domain.Guest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.byID[id]; ok {
		return g, nil
	}
	return nil, domain.ErrNotFound
}

func (r *MemoryGuestRepository) GetByEmail(_ context.Context, email string) (*domain.Guest, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if g, ok := r.byEmail[email]; ok {
		return g, nil
	}
	return nil, domain.ErrNotFound
}

func (r *MemoryGuestRepository) Create(_ context.Context, g *domain.Guest) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byEmail[g.Email]; ok {
		return 0, domain.ErrConflict
	}
	r.nextID++
	g.ID = r.nextID
	r.list = append(r.list, g)
	r.byID[g.ID] = g
	r.byEmail[g.Email] = g
	return g.ID, nil
}

// Count возвращает число гостей.
func (r *MemoryGuestRepository) Count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.list)
}
