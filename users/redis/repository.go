package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"go-rest-api/users"

	"github.com/redis/go-redis/v9"
)

// ttl is how long cached reads stay valid before falling back to the store.
const ttl = 5 * time.Minute

// Repository decorates another users.UserRepository with a Redis read
// cache for GetByID. Login/Register are never cached: they carry
// credentials and must always hit the real store.
type Repository struct {
	userRepository users.UserRepository
	redis          *redis.Client
}

var _ users.UserRepository = (*Repository)(nil)

func NewRepository(userRepository users.UserRepository, redis *redis.Client) *Repository {
	return &Repository{userRepository: userRepository, redis: redis}
}

func (r *Repository) Register(user *users.User) error {
	return r.userRepository.Register(user)
}

func (r *Repository) Login(user *users.User) error {
	return r.userRepository.Login(user)
}

func (r *Repository) GetByID(id int64) (*users.User, error) {
	ctx := context.Background()
	key := userKey(id)

	if cached, err := r.redis.Get(ctx, key).Bytes(); err == nil {
		var u users.User
		if err := json.Unmarshal(cached, &u); err == nil {
			return &u, nil
		}
	}

	user, err := r.userRepository.GetByID(id)
	if err != nil {
		return nil, err
	}
	if data, err := json.Marshal(user); err == nil {
		r.redis.Set(ctx, key, data, ttl)
	}
	return user, nil
}

func (r *Repository) Update(user *users.User) error {
	if err := r.userRepository.Update(user); err != nil {
		return err
	}
	r.redis.Del(context.Background(), userKey(user.ID))
	return nil
}

func (r *Repository) Delete(id int64) error {
	if err := r.userRepository.Delete(id); err != nil {
		return err
	}
	r.redis.Del(context.Background(), userKey(id))
	return nil
}

func userKey(id int64) string {
	return fmt.Sprintf("users:%d", id)
}
