package sqlite

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/kumarutkarsh99/students-api/internal/config"
	"github.com/kumarutkarsh99/students-api/internal/storage"
)

func TestStudentCRUD(t *testing.T) {
	store, err := New(&config.Config{StoragePath: filepath.Join(t.TempDir(), "students.db")})
	if err != nil {
		t.Fatalf("create sqlite storage: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close sqlite storage: %v", err)
		}
	})

	created, err := store.CreateStudent("Ada Lovelace", "ada@example.com", 36)
	if err != nil {
		t.Fatalf("create student: %v", err)
	}
	if created.ID <= 0 || created.Name != "Ada Lovelace" || created.Email != "ada@example.com" || created.Age != 36 {
		t.Fatalf("unexpected created student: %+v", created)
	}

	got, err := store.GetStudentById(created.ID)
	if err != nil {
		t.Fatalf("get student: %v", err)
	}
	if got != created {
		t.Fatalf("got student %+v, want %+v", got, created)
	}

	updated, err := store.UpdateStudentById(created.ID, "Ada Byron", "ada.b@example.com", 37)
	if err != nil {
		t.Fatalf("update student: %v", err)
	}
	want := created
	want.Name = "Ada Byron"
	want.Email = "ada.b@example.com"
	want.Age = 37
	if updated != want {
		t.Fatalf("updated student %+v, want %+v", updated, want)
	}

	students, err := store.GetStudents()
	if err != nil {
		t.Fatalf("list students: %v", err)
	}
	if len(students) != 1 || students[0] != want {
		t.Fatalf("listed students %+v, want [%+v]", students, want)
	}

	if err := store.DeleteStudentById(created.ID); err != nil {
		t.Fatalf("delete student: %v", err)
	}
	if _, err := store.GetStudentById(created.ID); !errors.Is(err, storage.ErrStudentNotFound) {
		t.Fatalf("get deleted student error = %v, want ErrStudentNotFound", err)
	}
}

func TestMissingStudentOperationsReturnNotFound(t *testing.T) {
	store, err := New(&config.Config{StoragePath: filepath.Join(t.TempDir(), "students.db")})
	if err != nil {
		t.Fatalf("create sqlite storage: %v", err)
	}
	t.Cleanup(func() {
		if err := store.Close(); err != nil {
			t.Errorf("close sqlite storage: %v", err)
		}
	})

	if _, err := store.GetStudentById(42); !errors.Is(err, storage.ErrStudentNotFound) {
		t.Errorf("get error = %v, want ErrStudentNotFound", err)
	}
	if _, err := store.UpdateStudentById(42, "Nobody", "nobody@example.com", 20); !errors.Is(err, storage.ErrStudentNotFound) {
		t.Errorf("update error = %v, want ErrStudentNotFound", err)
	}
	if err := store.DeleteStudentById(42); !errors.Is(err, storage.ErrStudentNotFound) {
		t.Errorf("delete error = %v, want ErrStudentNotFound", err)
	}
}
