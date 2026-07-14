package events

import (
	"errors"
	"testing"
)

// mockRepository implements EventRepository with per-call function fields,
// so each test only wires up the behavior it cares about.
type mockRepository struct {
	createFunc  func(*Event) error
	getAllFunc  func() ([]Event, error)
	getByIDFunc func(int64) (*Event, error)
	updateFunc  func(*Event) error
	deleteFunc  func(int64) error
}

var _ EventRepository = (*mockRepository)(nil)

func (m *mockRepository) Create(event *Event) error        { return m.createFunc(event) }
func (m *mockRepository) GetAll() ([]Event, error)         { return m.getAllFunc() }
func (m *mockRepository) GetByID(id int64) (*Event, error) { return m.getByIDFunc(id) }
func (m *mockRepository) Update(event *Event) error        { return m.updateFunc(event) }
func (m *mockRepository) Delete(id int64) error            { return m.deleteFunc(id) }

func TestServiceCreateDelegatesToRepository(t *testing.T) {
	event := &Event{Name: "Launch"}
	var got *Event
	repo := &mockRepository{createFunc: func(e *Event) error {
		got = e
		e.ID = 7
		return nil
	}}

	if err := NewService(repo).Create(event); err != nil {
		t.Fatalf("Create: %v", err)
	}
	if got != event {
		t.Error("Create did not pass the same *Event through to the repository")
	}
	if event.ID != 7 {
		t.Errorf("event.ID = %d, want 7 (repository-assigned id should be visible to the caller)", event.ID)
	}
}

func TestServiceCreatePropagatesError(t *testing.T) {
	wantErr := errors.New("insert failed")
	repo := &mockRepository{createFunc: func(*Event) error { return wantErr }}

	if err := NewService(repo).Create(&Event{}); !errors.Is(err, wantErr) {
		t.Errorf("Create error = %v, want %v", err, wantErr)
	}
}

func TestServiceList(t *testing.T) {
	want := []Event{{ID: 1}, {ID: 2}}
	repo := &mockRepository{getAllFunc: func() ([]Event, error) { return want, nil }}

	got, err := NewService(repo).List()
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(got) != len(want) {
		t.Fatalf("List returned %d events, want %d", len(got), len(want))
	}
}

func TestServiceGetByIDNotFound(t *testing.T) {
	repo := &mockRepository{getByIDFunc: func(int64) (*Event, error) { return nil, ErrEventNotFound }}

	_, err := NewService(repo).GetByID(99)
	if !errors.Is(err, ErrEventNotFound) {
		t.Errorf("GetByID error = %v, want %v", err, ErrEventNotFound)
	}
}

func TestServiceUpdateDelegatesToRepository(t *testing.T) {
	event := &Event{ID: 5, Name: "Updated"}
	var got *Event
	repo := &mockRepository{updateFunc: func(e *Event) error {
		got = e
		return nil
	}}

	if err := NewService(repo).Update(event); err != nil {
		t.Fatalf("Update: %v", err)
	}
	if got != event {
		t.Error("Update did not pass the same *Event through to the repository")
	}
}

func TestServiceDeleteDelegatesToRepository(t *testing.T) {
	var gotID int64
	repo := &mockRepository{deleteFunc: func(id int64) error {
		gotID = id
		return nil
	}}

	if err := NewService(repo).Delete(5); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	if gotID != 5 {
		t.Errorf("Delete called repository with id %d, want 5", gotID)
	}
}
