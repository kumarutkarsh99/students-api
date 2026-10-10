package storage

import (
	"errors"

	"github.com/kumarutkarsh99/students-api/internal/types"
)

var ErrStudentNotFound = errors.New("student not found")

type Storage interface {
	CreateStudent(name string, email string, age int) (types.Student, error)
	GetStudentById(id int64) (types.Student, error)
	GetStudents() ([]types.Student, error)
	UpdateStudentById(id int64, name string, email string, age int) (types.Student, error)
	DeleteStudentById(id int64) error
}
