package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-rest-api/events"

	"github.com/redis/go-redis/v9"
)

// ttl is how long cached reads stay valid before falling back to the store.
const ttl = 5 * time.Minute

const allKey = "events:all"

// Repository decorates another events.EventRepository with a Redis read
// cache. Reads are served from Redis when present; writes go straight to
// the wrapped repository and then invalidate the relevant cache entries.
type Repository struct {
	eventRepository events.EventRepository
	redis           *redis.Client
}

var _ events.EventRepository = (*Repository)(nil)

func NewRepository(eventRepository events.EventRepository, redis *redis.Client) *Repository {
	return &Repository{eventRepository: eventRepository, redis: redis}
}

func (r *Repository) Create(event *events.Event) error {
	if err := r.eventRepository.Create(event); err != nil {
		return err
	}
	r.redis.Del(context.Background(), allKey)
	return nil
}

func (r *Repository) GetAll() ([]events.Event, error) {
	ctx := context.Background()

	if cached, err := r.redis.Get(ctx, allKey).Bytes(); err == nil {
		var result []events.Event
		if err := json.Unmarshal(cached, &result); err == nil {
			return result, nil
		}
	}

	result, err := r.eventRepository.GetAll()
	if err != nil {
		return nil, err
	}
	if data, err := json.Marshal(result); err == nil {
		r.redis.Set(ctx, allKey, data, ttl)
	}
	return result, nil
}

func (r *Repository) GetByID(id int64) (*events.Event, error) {
	ctx := context.Background()
	key := eventKey(id)

	if cached, err := r.redis.Get(ctx, key).Bytes(); err == nil {
		var e events.Event
		if err := json.Unmarshal(cached, &e); err == nil {
			return &e, nil
		}
	}

	event, err := r.eventRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if data, err := json.Marshal(event); err == nil {
		r.redis.Set(ctx, key, data, ttl)
	}
	return event, nil
}

func (r *Repository) Update(event *events.Event) error {
	if err := r.eventRepository.Update(event); err != nil {
		return err
	}
	r.invalidate(event.ID)
	return nil
}

func (r *Repository) Delete(id int64) error {
	if err := r.eventRepository.Delete(id); err != nil {
		return err
	}
	r.invalidate(id)
	return nil
}

func (r *Repository) RegisterUserToEvent(eventID, userID int64) error {
	if err := r.eventRepository.RegisterUserToEvent(eventID, userID); err != nil {
		return err
	}
	r.invalidate(eventID)
	return nil
}

func (r *Repository) UnregisterUserFromEvent(eventID, userID int64) error {
	if err := r.eventRepository.UnregisterUserFromEvent(eventID, userID); err != nil {
		return err
	}
	r.invalidate(eventID)
	return nil
}

func (r *Repository) invalidate(id int64) {
	ctx := context.Background()
	r.redis.Del(ctx, eventKey(id), allKey)
}

func eventKey(id int64) string {
	return fmt.Sprintf("events:%d", id)
}
