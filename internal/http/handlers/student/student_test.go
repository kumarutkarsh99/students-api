package student

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kumarutkarsh99/students-api/internal/storage"
	"github.com/kumarutkarsh99/students-api/internal/types"
)

type mockStorage struct {
	created       types.Student
	got           types.Student
	list          []types.Student
	updated       types.Student
	getErr        error
	listErr       error
	createErr     error
	updateErr     error
	deleteErr     error
	lastUpdatedID int64
	lastDeletedID int64
}

func (m *mockStorage) CreateStudent(name, email string, age int) (types.Student, error) {
	if m.createErr != nil {
		return types.Student{}, m.createErr
	}
	m.created = types.Student{ID: 5, Name: name, Email: email, Age: age}
	return m.created, nil
}

func (m *mockStorage) GetStudentById(id int64) (types.Student, error) {
	if m.getErr != nil {
		return types.Student{}, m.getErr
	}
	return m.got, nil
}

func (m *mockStorage) GetStudents() ([]types.Student, error) {
	return m.list, m.listErr
}

func (m *mockStorage) UpdateStudentById(id int64, name, email string, age int) (types.Student, error) {
	m.lastUpdatedID = id
	if m.updateErr != nil {
		return types.Student{}, m.updateErr
	}
	m.updated = types.Student{ID: id, Name: name, Email: email, Age: age}
	return m.updated, nil
}

func (m *mockStorage) DeleteStudentById(id int64) error {
	m.lastDeletedID = id
	return m.deleteErr
}

func TestCreateStudent(t *testing.T) {
	store := &mockStorage{}
	req := httptest.NewRequest(http.MethodPost, "/api/students", strings.NewReader(
		`{"name":"Ada Lovelace","email":"ada@example.com","age":36}`,
	))
	rec := httptest.NewRecorder()

	New(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusCreated, rec.Body)
	}
	if got := rec.Header().Get("Location"); got != "/api/students/5" {
		t.Errorf("Location = %q, want %q", got, "/api/students/5")
	}
	if store.created.ID != 5 || store.created.Name != "Ada Lovelace" {
		t.Errorf("created student = %+v", store.created)
	}
}

func TestUpdateUsesPathIDAndReturnsUpdatedStudent(t *testing.T) {
	store := &mockStorage{}
	req := httptest.NewRequest(http.MethodPut, "/api/students/17", strings.NewReader(
		`{"name":"Ada Byron","email":"ada@example.com","age":37}`,
	))
	req.SetPathValue("id", "17")
	rec := httptest.NewRecorder()

	UpdateStudentById(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d; body: %s", rec.Code, http.StatusOK, rec.Body)
	}
	if store.lastUpdatedID != 17 || store.updated.ID != 17 {
		t.Errorf("updated ID = %d, want path ID 17", store.lastUpdatedID)
	}
}

func TestUpdateRejectsIDInRequestBody(t *testing.T) {
	store := &mockStorage{}
	req := httptest.NewRequest(http.MethodPut, "/api/students/17", strings.NewReader(
		`{"id":99,"name":"Ada Byron","email":"ada@example.com","age":37}`,
	))
	req.SetPathValue("id", "17")
	rec := httptest.NewRecorder()

	UpdateStudentById(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if store.lastUpdatedID != 0 {
		t.Error("storage update called for a body containing an unknown ID field")
	}
}

func TestHandlersRejectInvalidIDs(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{name: "not a number", id: "abc"},
		{name: "zero", id: "0"},
		{name: "negative", id: "-1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/students/"+tt.id, nil)
			req.SetPathValue("id", tt.id)
			rec := httptest.NewRecorder()

			GetStudentById(&mockStorage{}).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
		})
	}
}

func TestGetStudentNotFound(t *testing.T) {
	store := &mockStorage{getErr: storage.ErrStudentNotFound}
	req := httptest.NewRequest(http.MethodGet, "/api/students/17", nil)
	req.SetPathValue("id", "17")
	rec := httptest.NewRecorder()

	GetStudentById(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestUpdateStudentNotFound(t *testing.T) {
	store := &mockStorage{updateErr: storage.ErrStudentNotFound}
	req := httptest.NewRequest(http.MethodPut, "/api/students/17", strings.NewReader(
		`{"name":"Ada Byron","email":"ada@example.com","age":37}`,
	))
	req.SetPathValue("id", "17")
	rec := httptest.NewRecorder()

	UpdateStudentById(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestDeleteStudent(t *testing.T) {
	store := &mockStorage{}
	req := httptest.NewRequest(http.MethodDelete, "/api/students/17", nil)
	req.SetPathValue("id", "17")
	rec := httptest.NewRecorder()

	DeleteStudentById(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNoContent)
	}
	if store.lastDeletedID != 17 {
		t.Errorf("deleted ID = %d, want 17", store.lastDeletedID)
	}
}

func TestDeleteStudentNotFound(t *testing.T) {
	store := &mockStorage{deleteErr: storage.ErrStudentNotFound}
	req := httptest.NewRequest(http.MethodDelete, "/api/students/17", nil)
	req.SetPathValue("id", "17")
	rec := httptest.NewRecorder()

	DeleteStudentById(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusNotFound)
	}
}

func TestListErrorDoesNotReturnSuccess(t *testing.T) {
	store := &mockStorage{listErr: errors.New("database unavailable")}
	rec := httptest.NewRecorder()

	GetList(store).ServeHTTP(rec, httptest.NewRequest(http.MethodGet, "/api/students", nil))

	if rec.Code != http.StatusInternalServerError {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusInternalServerError)
	}
}

func TestUpdateRejectsInvalidBody(t *testing.T) {
	store := &mockStorage{}
	req := httptest.NewRequest(http.MethodPut, "/api/students/17", strings.NewReader(
		`{"name":"Ada","email":"not-an-email","age":0}`,
	))
	req.SetPathValue("id", "17")
	rec := httptest.NewRecorder()

	UpdateStudentById(store).ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
	}
	if store.lastUpdatedID != 0 {
		t.Error("storage update called for invalid input")
	}
}

func TestCreateRejectsEmptyAndMalformedBodies(t *testing.T) {
	tests := []struct {
		name string
		body string
	}{
		{name: "empty", body: ""},
		{name: "malformed", body: `{"name":`},
		{name: "multiple values", body: `{"name":"Ada","email":"ada@example.com","age":36} {}`},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			store := &mockStorage{}
			req := httptest.NewRequest(http.MethodPost, "/api/students", strings.NewReader(tt.body))
			rec := httptest.NewRecorder()

			New(store).ServeHTTP(rec, req)

			if rec.Code != http.StatusBadRequest {
				t.Errorf("status = %d, want %d", rec.Code, http.StatusBadRequest)
			}
			if store.created.ID != 0 {
				t.Error("storage create called for invalid input")
			}
		})
	}
}
