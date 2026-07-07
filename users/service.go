package users

// Service holds the application use-cases for events. It depends on the
// EventRepository abstraction, so the concrete storage can be swapped or mocked.
type Service struct {
	repo UserRepository
}

func NewService(repo UserRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Register(user *User) error {
	return s.repo.Register(user)
}

func (s *Service) Login(user *User) error {
	return s.repo.Login(user)
}

func (s *Service) GetByID(id int64) (*User, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Update(user *User) error {
	return s.repo.Update(user)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
