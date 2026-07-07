package events

// Service holds the application use-cases for events. It depends on the
// EventRepository abstraction, so the concrete storage can be swapped or mocked.
type Service struct {
	repo EventRepository
}

func NewService(repo EventRepository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(event *Event) error {
	return s.repo.Create(event)
}

func (s *Service) List() ([]Event, error) {
	return s.repo.GetAll()
}

func (s *Service) GetByID(id int64) (*Event, error) {
	return s.repo.GetByID(id)
}

func (s *Service) Update(event *Event) error {
	return s.repo.Update(event)
}

func (s *Service) Delete(id int64) error {
	return s.repo.Delete(id)
}
