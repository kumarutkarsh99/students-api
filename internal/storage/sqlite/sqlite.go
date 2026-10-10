package sqlite

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kumarutkarsh99/students-api/internal/config"
	"github.com/kumarutkarsh99/students-api/internal/storage"
	"github.com/kumarutkarsh99/students-api/internal/types"
	_ "modernc.org/sqlite"
)

type Sqlite struct {
	db *sql.DB
}

var _ storage.Storage = (*Sqlite)(nil)

func New(cfg *config.Config) (*Sqlite, error) {
	if cfg.StoragePath != ":memory:" {
		if err := os.MkdirAll(filepath.Dir(cfg.StoragePath), 0o755); err != nil {
			return nil, fmt.Errorf("create storage directory: %w", err)
		}
	}

	db, err := sql.Open("sqlite", cfg.StoragePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("connect to sqlite database: %w", err)
	}
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS students(
	id INTEGER PRIMARY KEY AUTOINCREMENT,
	name TEXT NOT NULL,
	email TEXT NOT NULL,
	age INTEGER NOT NULL CHECK (age > 0)
	)`)

	if err != nil {
		db.Close()
		return nil, fmt.Errorf("create students table: %w", err)
	}

	return &Sqlite{
		db: db,
	}, nil
}

func (s *Sqlite) Close() error {
	return s.db.Close()
}

func (s *Sqlite) CreateStudent(name string, email string, age int) (types.Student, error) {
	res, err := s.db.Exec("INSERT INTO students (name, email, age) VALUES (?, ?, ?)", name, email, age)
	if err != nil {
		return types.Student{}, fmt.Errorf("insert student: %w", err)
	}

	id, err := res.LastInsertId()
	if err != nil {
		return types.Student{}, fmt.Errorf("get inserted student id: %w", err)
	}

	return types.Student{ID: id, Name: name, Email: email, Age: age}, nil
}

func (s *Sqlite) GetStudentById(id int64) (types.Student, error) {
	var student types.Student

	err := s.db.QueryRow("SELECT id, name, email, age FROM students WHERE id = ?", id).
		Scan(&student.ID, &student.Name, &student.Email, &student.Age)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return types.Student{}, fmt.Errorf("%w: id %d", storage.ErrStudentNotFound, id)
		}
		return types.Student{}, fmt.Errorf("get student with id %d: %w", id, err)
	}

	return student, nil
}

func (s *Sqlite) GetStudents() ([]types.Student, error) {
	rows, err := s.db.Query("SELECT id, name, email, age FROM students ORDER BY id")
	if err != nil {
		return nil, fmt.Errorf("list students: %w", err)
	}
	defer rows.Close()

	students := make([]types.Student, 0)
	for rows.Next() {
		var student types.Student
		err := rows.Scan(&student.ID, &student.Name, &student.Email, &student.Age)
		if err != nil {
			return nil, fmt.Errorf("iterate students: %w", err)
		}
		students = append(students, student)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate students: %w", err)
	}
	return students, nil
}

func (s *Sqlite) DeleteStudentById(id int64) error {
	res, err := s.db.Exec("DELETE FROM students WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("delete student with id %d: %w", id, err)
	}

	rowsAffected, err := res.RowsAffected()
	if err != nil {
		return fmt.Errorf("check rows affected when deleting student with id %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return fmt.Errorf("%w: id %d", storage.ErrStudentNotFound, id)
	}

	return nil
}

func (s *Sqlite) UpdateStudentById(id int64, name string, email string, age int) (types.Student, error) {
	stmt, err := s.db.Exec("UPDATE students SET name = ?, email = ?, age = ? WHERE id = ?", name, email, age, id)
	if err != nil {
		return types.Student{}, fmt.Errorf("update student with id %d: %w", id, err)
	}

	rowsAffected, err := stmt.RowsAffected()
	if err != nil {
		return types.Student{}, fmt.Errorf("check rows affected when updating student with id %d: %w", id, err)
	}
	if rowsAffected == 0 {
		return types.Student{}, fmt.Errorf("%w: id %d", storage.ErrStudentNotFound, id)
	}
	return types.Student{ID: id, Name: name, Email: email, Age: age}, nil
}
