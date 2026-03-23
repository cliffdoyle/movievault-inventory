package repository

import (
	"regexp"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/cliffdoyle/eks-golang/internal/models"
)

func TestMovieRepository_GetMovies(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &movieRepository{db: db}

	rows := sqlmock.NewRows([]string{"id", "title", "genre", "year", "stock"}).
		AddRow(1, "The Matrix", "Sci-Fi", 1999, 10).
		AddRow(2, "Inception", "Sci-Fi", 2010, 5)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, genre, year, stock FROM movies ORDER BY id")).
		WillReturnRows(rows)

	movies, err := repo.GetMovies()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(movies) != 2 {
		t.Errorf("expected 2 movies, got %d", len(movies))
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestMovieRepository_CreateMovie(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &movieRepository{db: db}

	movie := &models.Movie{
		Title: "The Matrix",
		Genre: "Sci-Fi",
		Year:  1999,
		Stock: 10,
	}

	rows := sqlmock.NewRows([]string{"id"}).AddRow(1)
	mock.ExpectQuery(regexp.QuoteMeta("INSERT INTO movies(title, genre, year, stock) VALUES($1, $2, $3, $4) RETURNING id")).
		WithArgs(movie.Title, movie.Genre, movie.Year, movie.Stock).
		WillReturnRows(rows)

	err = repo.CreateMovie(movie)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if movie.ID != 1 {
		t.Errorf("expected ID 1, got %d", movie.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestMovieRepository_GetMovieByID(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &movieRepository{db: db}

	rows := sqlmock.NewRows([]string{"id", "title", "genre", "year", "stock"}).
		AddRow(1, "The Matrix", "Sci-Fi", 1999, 10)

	mock.ExpectQuery(regexp.QuoteMeta("SELECT id, title, genre, year, stock FROM movies WHERE id=$1")).
		WithArgs(1).
		WillReturnRows(rows)

	movie, err := repo.GetMovieByID(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if movie.ID != 1 {
		t.Errorf("expected ID 1, got %d", movie.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestMovieRepository_UpdateStock(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &movieRepository{db: db}

	rows := sqlmock.NewRows([]string{"id", "title", "genre", "year", "stock"}).
		AddRow(1, "The Matrix", "Sci-Fi", 1999, 15)

	mock.ExpectQuery(regexp.QuoteMeta("UPDATE movies SET stock=$1 WHERE id=$2 RETURNING id, title, genre, year, stock")).
		WithArgs(15, 1).
		WillReturnRows(rows)

	movie, err := repo.UpdateStock(1, 15)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if movie.Stock != 15 {
		t.Errorf("expected stock 15, got %d", movie.Stock)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestMovieRepository_DeleteMovie(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &movieRepository{db: db}

	rows := sqlmock.NewRows([]string{"id", "title", "genre", "year", "stock"}).
		AddRow(1, "The Matrix", "Sci-Fi", 1999, 10)

	mock.ExpectQuery(regexp.QuoteMeta("DELETE FROM movies WHERE id=$1 RETURNING id, title, genre, year, stock")).
		WithArgs(1).
		WillReturnRows(rows)

	movie, err := repo.DeleteMovie(1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if movie.ID != 1 {
		t.Errorf("expected ID 1, got %d", movie.ID)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}

func TestMovieRepository_Ping(t *testing.T) {
	db, mock, err := sqlmock.New(sqlmock.MonitorPingsOption(true))
	if err != nil {
		t.Fatalf("failed to create sqlmock: %v", err)
	}
	defer db.Close()

	repo := &movieRepository{db: db}

	mock.ExpectPing()

	err = repo.Ping()
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Errorf("unfulfilled expectations: %s", err)
	}
}
