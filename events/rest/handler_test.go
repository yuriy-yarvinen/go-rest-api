package rest

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"go-rest-api/authctx"
	"go-rest-api/events"

	"github.com/gin-gonic/gin"
)

type mockRepository struct {
	createFunc  func(*events.Event) error
	getAllFunc  func() ([]events.Event, error)
	getByIDFunc func(int64) (*events.Event, error)
	updateFunc  func(*events.Event) error
	deleteFunc  func(int64) error
}

var _ events.EventRepository = (*mockRepository)(nil)

func (m *mockRepository) Create(event *events.Event) error        { return m.createFunc(event) }
func (m *mockRepository) GetAll() ([]events.Event, error)         { return m.getAllFunc() }
func (m *mockRepository) GetByID(id int64) (*events.Event, error) { return m.getByIDFunc(id) }
func (m *mockRepository) Update(event *events.Event) error        { return m.updateFunc(event) }
func (m *mockRepository) Delete(id int64) error                   { return m.deleteFunc(id) }

// fakeAuth stands in for users/rest.AuthRequired in these handler tests: it
// just stamps the given userID onto the context via authctx, so the
// ownership logic in the handlers can be tested without a real JWT/user
// lookup round trip (that flow is already covered by
// users/rest/middleware_test.go).
func fakeAuth(userID int64) gin.HandlerFunc {
	return func(c *gin.Context) {
		authctx.SetUserID(c, userID)
		c.Next()
	}
}

func newTestRouter(h *Handler, authUserID int64) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	RegisterRoutes(router.Group("/"), h, fakeAuth(authUserID))
	return router
}

func doJSONRequest(router *gin.Engine, method, path string, body any) *httptest.ResponseRecorder {
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			panic(err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestCreateSetsOwnerFromToken(t *testing.T) {
	var got *events.Event
	repo := &mockRepository{createFunc: func(e *events.Event) error {
		got = e
		return nil
	}}
	router := newTestRouter(NewHandler(events.NewService(repo)), 42)

	// Client tries to claim the event for user 999; the server must ignore
	// that and force the authenticated caller's id instead.
	rec := doJSONRequest(router, http.MethodPost, "/events", map[string]any{
		"name": "Launch", "location": "Remote", "user_id": 999,
	})

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusCreated, rec.Body.String())
	}
	if got == nil {
		t.Fatal("repository Create was never called")
	}
	if got.UserID != 42 {
		t.Errorf("event.UserID = %d, want 42 (the authenticated user, not the client-supplied 999)", got.UserID)
	}
}

func TestUpdateRejectsNonOwner(t *testing.T) {
	repo := &mockRepository{
		getByIDFunc: func(id int64) (*events.Event, error) {
			return &events.Event{ID: id, UserID: 1}, nil
		},
		updateFunc: func(*events.Event) error {
			t.Fatal("Update should not be called when the caller doesn't own the event")
			return nil
		},
	}
	router := newTestRouter(NewHandler(events.NewService(repo)), 2) // different user than the owner

	rec := doJSONRequest(router, http.MethodPut, "/events/1", map[string]any{
		"name": "Hijacked", "location": "Remote",
	})

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestUpdateAllowsOwnerAndPreservesUserID(t *testing.T) {
	var got *events.Event
	repo := &mockRepository{
		getByIDFunc: func(id int64) (*events.Event, error) {
			return &events.Event{ID: id, UserID: 1}, nil
		},
		updateFunc: func(e *events.Event) error {
			got = e
			return nil
		},
	}
	router := newTestRouter(NewHandler(events.NewService(repo)), 1) // the owner

	// Client tries to reassign the event to user 999 via the body; the
	// server must keep the original owner regardless.
	rec := doJSONRequest(router, http.MethodPut, "/events/1", map[string]any{
		"name": "Updated", "location": "Remote", "user_id": 999,
	})

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if got == nil {
		t.Fatal("repository Update was never called")
	}
	if got.UserID != 1 {
		t.Errorf("event.UserID = %d, want 1 (ownership must not change via update)", got.UserID)
	}
}

func TestDeleteRejectsNonOwner(t *testing.T) {
	repo := &mockRepository{
		getByIDFunc: func(id int64) (*events.Event, error) {
			return &events.Event{ID: id, UserID: 1}, nil
		},
		deleteFunc: func(int64) error {
			t.Fatal("Delete should not be called when the caller doesn't own the event")
			return nil
		},
	}
	router := newTestRouter(NewHandler(events.NewService(repo)), 2)

	rec := doJSONRequest(router, http.MethodDelete, "/events/1", nil)

	if rec.Code != http.StatusForbidden {
		t.Errorf("status = %d, want %d, body=%s", rec.Code, http.StatusForbidden, rec.Body.String())
	}
}

func TestDeleteAllowsOwner(t *testing.T) {
	deleted := false
	repo := &mockRepository{
		getByIDFunc: func(id int64) (*events.Event, error) {
			return &events.Event{ID: id, UserID: 1}, nil
		},
		deleteFunc: func(int64) error {
			deleted = true
			return nil
		},
	}
	router := newTestRouter(NewHandler(events.NewService(repo)), 1)

	rec := doJSONRequest(router, http.MethodDelete, "/events/1", nil)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !deleted {
		t.Error("repository Delete was never called")
	}
}
