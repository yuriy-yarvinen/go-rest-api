package rest

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"go-rest-api/users"
	"go-rest-api/utils"

	"github.com/gin-gonic/gin"
)

type mockUserRepository struct {
	getByIDFunc func(int64) (*users.User, error)
}

var _ users.UserRepository = (*mockUserRepository)(nil)

func (m *mockUserRepository) Register(*users.User) error            { return nil }
func (m *mockUserRepository) Login(*users.User) error               { return nil }
func (m *mockUserRepository) GetByID(id int64) (*users.User, error) { return m.getByIDFunc(id) }
func (m *mockUserRepository) Update(*users.User) error              { return nil }
func (m *mockUserRepository) Delete(int64) error                    { return nil }

// newTestRouter wires AuthRequired in front of a dummy handler that echoes
// back the userID the middleware put in the context, so tests can assert
// on both the status code and what got passed through.
func newTestRouter(service *users.Service) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/protected", AuthRequired(service), func(c *gin.Context) {
		userID, ok := UserID(c)
		if !ok {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "no userID in context"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"userID": userID})
	})
	return router
}

func doRequest(router *gin.Engine, authHeader string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	if authHeader != "" {
		req.Header.Set("Authorization", authHeader)
	}
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)
	return rec
}

func TestAuthRequiredMissingHeader(t *testing.T) {
	router := newTestRouter(users.NewService(&mockUserRepository{}))

	rec := doRequest(router, "")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequiredMalformedHeader(t *testing.T) {
	router := newTestRouter(users.NewService(&mockUserRepository{}))

	rec := doRequest(router, "NotBearer sometoken")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequiredInvalidToken(t *testing.T) {
	router := newTestRouter(users.NewService(&mockUserRepository{}))

	rec := doRequest(router, "Bearer garbage.token.here")
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequiredRejectsTokenForDeletedUser(t *testing.T) {
	tokenString, err := utils.GenerateJWT("alice@gmail.com", 42)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	router := newTestRouter(users.NewService(&mockUserRepository{
		getByIDFunc: func(int64) (*users.User, error) { return nil, users.ErrUserNotFound },
	}))

	rec := doRequest(router, "Bearer "+tokenString)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("status = %d, want %d (a valid token for a deleted user must still be rejected)", rec.Code, http.StatusUnauthorized)
	}
}

func TestAuthRequiredValidTokenSetsUserID(t *testing.T) {
	tokenString, err := utils.GenerateJWT("alice@gmail.com", 42)
	if err != nil {
		t.Fatalf("GenerateJWT: %v", err)
	}

	router := newTestRouter(users.NewService(&mockUserRepository{
		getByIDFunc: func(id int64) (*users.User, error) { return &users.User{ID: id}, nil },
	}))

	rec := doRequest(router, "Bearer "+tokenString)
	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d, body=%s", rec.Code, http.StatusOK, rec.Body.String())
	}
	if !strings.Contains(rec.Body.String(), `"userID":42`) {
		t.Errorf("body = %s, want it to contain userID 42", rec.Body.String())
	}
}
