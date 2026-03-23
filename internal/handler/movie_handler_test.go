package handler

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/cliffdoyle/eks-golang/internal/models"
)

type mockMovieService struct {
	returnMovies []models.Movie
	returnMovie  models.Movie
	returnErr    error
}

func (m *mockMovieService) GetMovies() ([]models.Movie, error)          { return m.returnMovies, m.returnErr }
func (m *mockMovieService) CreateMovie(movie *models.Movie) error       { movie.ID = 1; return m.returnErr }
func (m *mockMovieService) GetMovieByID(id int) (models.Movie, error)   { return m.returnMovie, m.returnErr }
func (m *mockMovieService) UpdateStock(id int, stock int) (models.Movie, error) {
	m.returnMovie.Stock = stock
	return m.returnMovie, m.returnErr
}
func (m *mockMovieService) DeleteMovie(id int) (models.Movie, error)    { return m.returnMovie, m.returnErr }
func (m *mockMovieService) PingDB() error                               { return m.returnErr }

func TestHandleHealth_OK(t *testing.T) {
	svc := &mockMovieService{}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.HandleHealth(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHandleHealth_Unavailable(t *testing.T) {
	svc := &mockMovieService{returnErr: errors.New("db error")}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	rr := httptest.NewRecorder()
	h.HandleHealth(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("expected 503, got %d", rr.Code)
	}
}

func TestHandleMovies_Get(t *testing.T) {
	svc := &mockMovieService{
		returnMovies: []models.Movie{
			{ID: 1, Title: "The Matrix"},
		},
	}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/movies", nil)
	rr := httptest.NewRecorder()
	h.HandleMovies(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var movies []models.Movie
	json.NewDecoder(rr.Body).Decode(&movies)
	if len(movies) != 1 || movies[0].Title != "The Matrix" {
		t.Errorf("unexpected response: %s", rr.Body.String())
	}
}

func TestHandleMovies_Post(t *testing.T) {
	svc := &mockMovieService{}
	h := NewMovieHandler(svc)

	body := []byte(`{"title":"The Matrix","genre":"Sci-Fi","year":1999,"stock":10}`)
	req, _ := http.NewRequest(http.MethodPost, "/movies", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.HandleMovies(rr, req)

	if rr.Code != http.StatusCreated {
		t.Errorf("expected 201, got %d", rr.Code)
	}

	var m models.Movie
	json.NewDecoder(rr.Body).Decode(&m)
	if m.ID != 1 {
		t.Errorf("expected movie ID 1, got %d", m.ID)
	}
}

func TestHandleMovie_Get_Found(t *testing.T) {
	svc := &mockMovieService{
		returnMovie: models.Movie{ID: 1, Title: "The Matrix"},
	}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/movies/1", nil)
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}
}

func TestHandleMovie_Get_NotFound(t *testing.T) {
	svc := &mockMovieService{
		returnErr: sql.ErrNoRows,
	}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/movies/999", nil)
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleMovie_Put(t *testing.T) {
	svc := &mockMovieService{
		returnMovie: models.Movie{ID: 1, Title: "The Matrix"},
	}
	h := NewMovieHandler(svc)

	body := []byte(`{"stock":15}`)
	req, _ := http.NewRequest(http.MethodPut, "/movies/1", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusOK {
		t.Errorf("expected 200, got %d", rr.Code)
	}

	var m models.Movie
	json.NewDecoder(rr.Body).Decode(&m)
	if m.Stock != 15 {
		t.Errorf("expected stock 15, got %d", m.Stock)
	}
}

func TestHandleMovie_Delete(t *testing.T) {
	svc := &mockMovieService{
		returnMovie: models.Movie{ID: 1, Title: "The Matrix"},
	}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodDelete, "/movies/1", nil)
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusNoContent {
		t.Errorf("expected 204, got %d", rr.Code)
	}
}

func TestHandleMovies_Get_Error(t *testing.T) {
	svc := &mockMovieService{returnErr: errors.New("db error")}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/movies", nil)
	rr := httptest.NewRecorder()
	h.HandleMovies(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestHandleMovies_Post_BadJSON(t *testing.T) {
	svc := &mockMovieService{}
	h := NewMovieHandler(svc)

	body := []byte(`{"title": "The Matrix", "year": "NINETEEN}`) // invalid json
	req, _ := http.NewRequest(http.MethodPost, "/movies", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.HandleMovies(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleMovies_Post_Error(t *testing.T) {
	svc := &mockMovieService{returnErr: errors.New("db error")}
	h := NewMovieHandler(svc)

	body := []byte(`{"title":"The Matrix","genre":"Sci-Fi","year":1999,"stock":10}`)
	req, _ := http.NewRequest(http.MethodPost, "/movies", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.HandleMovies(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestHandleMovies_BadMethod(t *testing.T) {
	svc := &mockMovieService{}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodPatch, "/movies", nil)
	rr := httptest.NewRecorder()
	h.HandleMovies(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}

func TestHandleMovie_InvalidID(t *testing.T) {
	svc := &mockMovieService{}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/movies/abc", nil)
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleMovie_Get_InternalError(t *testing.T) {
	svc := &mockMovieService{returnErr: errors.New("db error")}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodGet, "/movies/1", nil)
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestHandleMovie_Put_BadJSON(t *testing.T) {
	svc := &mockMovieService{}
	h := NewMovieHandler(svc)

	body := []byte(`{"stock": "fifteen"}`) // invalid json
	req, _ := http.NewRequest(http.MethodPut, "/movies/1", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rr.Code)
	}
}

func TestHandleMovie_Put_NotFound(t *testing.T) {
	svc := &mockMovieService{returnErr: sql.ErrNoRows}
	h := NewMovieHandler(svc)

	body := []byte(`{"stock": 15}`)
	req, _ := http.NewRequest(http.MethodPut, "/movies/1", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleMovie_Put_Error(t *testing.T) {
	svc := &mockMovieService{returnErr: errors.New("db error")}
	h := NewMovieHandler(svc)

	body := []byte(`{"stock": 15}`)
	req, _ := http.NewRequest(http.MethodPut, "/movies/1", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestHandleMovie_Delete_NotFound(t *testing.T) {
	svc := &mockMovieService{returnErr: sql.ErrNoRows}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodDelete, "/movies/1", nil)
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusNotFound {
		t.Errorf("expected 404, got %d", rr.Code)
	}
}

func TestHandleMovie_Delete_Error(t *testing.T) {
	svc := &mockMovieService{returnErr: errors.New("db error")}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodDelete, "/movies/1", nil)
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", rr.Code)
	}
}

func TestHandleMovie_BadMethod(t *testing.T) {
	svc := &mockMovieService{}
	h := NewMovieHandler(svc)

	req, _ := http.NewRequest(http.MethodPatch, "/movies/1", nil)
	rr := httptest.NewRecorder()
	h.HandleMovie(rr, req)

	if rr.Code != http.StatusMethodNotAllowed {
		t.Errorf("expected 405, got %d", rr.Code)
	}
}
