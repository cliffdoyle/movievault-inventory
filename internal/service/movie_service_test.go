package service

import (
	"errors"
	"testing"

	"github.com/IBM/sarama"
	"github.com/IBM/sarama/mocks"
	"github.com/cliffdoyle/eks-golang/internal/models"
)

// mockMovieRepository is a stub for the MovieRepository interface.
type mockMovieRepository struct {
	returnMovies []models.Movie
	returnMovie  models.Movie
	returnErr    error
}

func (m *mockMovieRepository) GetMovies() ([]models.Movie, error) {
	return m.returnMovies, m.returnErr
}

func (m *mockMovieRepository) CreateMovie(movie *models.Movie) error {
	movie.ID = 1
	return m.returnErr
}

func (m *mockMovieRepository) GetMovieByID(id int) (models.Movie, error) {
	return m.returnMovie, m.returnErr
}

func (m *mockMovieRepository) UpdateStock(id int, stock int) (models.Movie, error) {
	m.returnMovie.Stock = stock
	return m.returnMovie, m.returnErr
}

func (m *mockMovieRepository) DeleteMovie(id int) (models.Movie, error) {
	return m.returnMovie, m.returnErr
}

func (m *mockMovieRepository) Ping() error {
	return m.returnErr
}

func TestMovieService_CreateMovie(t *testing.T) {
	repo := &mockMovieRepository{}
	
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer := mocks.NewSyncProducer(t, config)
	producer.ExpectSendMessageAndSucceed()

	svc := &movieService{
		repo:     repo,
		producer: producer,
	}

	movie := &models.Movie{
		Title: "The Matrix",
		Genre: "Sci-Fi",
		Year:  1999,
		Stock: 10,
	}

	err := svc.CreateMovie(movie)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if movie.ID != 1 {
		t.Errorf("expected movie ID 1, got %d", movie.ID)
	}
}

func TestMovieService_UpdateStock(t *testing.T) {
	repo := &mockMovieRepository{
		returnMovie: models.Movie{ID: 1, Title: "The Matrix", Stock: 10},
	}
	
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer := mocks.NewSyncProducer(t, config)
	producer.ExpectSendMessageAndSucceed()

	svc := &movieService{
		repo:     repo,
		producer: producer,
	}

	m, err := svc.UpdateStock(1, 15)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if m.Stock != 15 {
		t.Errorf("expected stock to be 15, got %d", m.Stock)
	}
}

func TestMovieService_DeleteMovie(t *testing.T) {
	repo := &mockMovieRepository{
		returnMovie: models.Movie{ID: 1, Title: "The Matrix"},
	}
	
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer := mocks.NewSyncProducer(t, config)
	producer.ExpectSendMessageAndSucceed()

	svc := &movieService{
		repo:     repo,
		producer: producer,
	}

	m, err := svc.DeleteMovie(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if m.ID != 1 {
		t.Errorf("expected ID 1, got %d", m.ID)
	}
}

func TestMovieService_GetMovies(t *testing.T) {
	repo := &mockMovieRepository{
		returnMovies: []models.Movie{
			{ID: 1, Title: "The Matrix"},
			{ID: 2, Title: "Inception"},
		},
	}
	
	svc := &movieService{
		repo:     repo,
		// producer not needed for GetMovies
	}

	movies, err := svc.GetMovies()
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if len(movies) != 2 {
		t.Errorf("expected 2 movies, got %d", len(movies))
	}
}

func TestMovieService_GetMovieByID(t *testing.T) {
	repo := &mockMovieRepository{
		returnMovie: models.Movie{ID: 1, Title: "The Matrix"},
	}
	
	svc := &movieService{
		repo:     repo,
	}

	m, err := svc.GetMovieByID(1)
	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	if m.ID != 1 {
		t.Errorf("expected ID 1, got %d", m.ID)
	}
}

func TestMovieService_CreateMovie_PropagatesError(t *testing.T) {
	// If the DB insert fails, the Kafka event shouldn't be published.
	repo := &mockMovieRepository{
		returnErr: errors.New("db error"),
	}
	
	config := sarama.NewConfig()
	config.Producer.Return.Successes = true
	producer := mocks.NewSyncProducer(t, config)
	// We do NOT expect SendMessage to be called.

	svc := &movieService{
		repo:     repo,
		producer: producer,
	}

	err := svc.CreateMovie(&models.Movie{Title: "The Matrix"})
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
}
