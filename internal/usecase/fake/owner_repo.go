package fake

import (
	"context"
	"sync"

	"github.com/ivan-rudev/ai-for-developers-project-387/internal/domain"
)

// MemoryOwnerRepository — in-memory OwnerRepository + OwnerProvisioner.
// Email уникален (как UNIQUE(owners.email)); нарушение → domain.ErrConflict.
type MemoryOwnerRepository struct {
	mu      sync.Mutex
	nextID  int64
	list    []*domain.Owner
	byID    map[int64]*domain.Owner
	byEmail map[string]*domain.Owner
	byUUID  map[string]*domain.Owner
	events  *MemoryEventRepository
}

func NewOwnerRepository() *MemoryOwnerRepository {
	return &MemoryOwnerRepository{
		byID:    make(map[int64]*domain.Owner),
		byEmail: make(map[string]*domain.Owner),
		byUUID:  make(map[string]*domain.Owner),
	}
}

func (r *MemoryOwnerRepository) GetAll(_ context.Context) ([]domain.Owner, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make([]domain.Owner, 0, len(r.list))
	for _, o := range r.list {
		out = append(out, *o)
	}
	return out, nil
}

func (r *MemoryOwnerRepository) GetByID(_ context.Context, id int64) (*domain.Owner, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.byID[id]; ok {
		return o, nil
	}
	return nil, domain.ErrNotFound
}

func (r *MemoryOwnerRepository) GetByUUID(_ context.Context, uuid string) (*domain.Owner, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.byUUID[uuid]; ok {
		return o, nil
	}
	return nil, domain.ErrNotFound
}

func (r *MemoryOwnerRepository) GetByEmail(_ context.Context, email string) (*domain.Owner, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if o, ok := r.byEmail[email]; ok {
		return o, nil
	}
	return nil, domain.ErrNotFound
}

func (r *MemoryOwnerRepository) Create(_ context.Context, o *domain.Owner) (int64, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if _, ok := r.byEmail[o.Email]; ok {
		return 0, domain.ErrConflict
	}
	r.nextID++
	o.ID = r.nextID
	r.list = append(r.list, o)
	r.byID[o.ID] = o
	r.byEmail[o.Email] = o
	r.byUUID[o.UUID] = o
	return o.ID, nil
}

func (r *MemoryOwnerRepository) Update(_ context.Context, o *domain.Owner) error {
	r.mu.Lock()
	defer r.mu.Unlock()
	existing, ok := r.byID[o.ID]
	if !ok {
		return domain.ErrNotFound
	}
	if existing.Email != o.Email {
		if other, exists := r.byEmail[o.Email]; exists && other.ID != o.ID {
			return domain.ErrConflict
		}
		delete(r.byEmail, existing.Email)
		r.byEmail[o.Email] = o
	}
	if existing.UUID != o.UUID {
		delete(r.byUUID, existing.UUID)
		r.byUUID[o.UUID] = o
	}
	*existing = *o
	return nil
}

// CreateOwnerWithDefaultEvents атомарно создаёт владельца и его события,
// имитируя транзакцию: при сбое создания события владелец откатывается.
func (r *MemoryOwnerRepository) CreateOwnerWithDefaultEvents(ctx context.Context, o *domain.Owner, events []domain.Event) (int64, error) {
	if _, err := r.Create(ctx, o); err != nil {
		return 0, err
	}
	for i := range events {
		events[i].OwnerID = o.ID
		if _, err := r.events.Create(ctx, &events[i]); err != nil {
			r.mu.Lock()
			delete(r.byID, o.ID)
			delete(r.byEmail, o.Email)
			delete(r.byUUID, o.UUID)
			r.mu.Unlock()
			return 0, err
		}
	}
	return o.ID, nil
}
