package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-playground/validator/v10"
	"github.com/kumarutkarsh99/students-api/internal/storage"
	"github.com/kumarutkarsh99/students-api/internal/types"
	"github.com/kumarutkarsh99/students-api/internal/utils/response"
)

const maxRequestBodyBytes = 1 << 20

var validate = validator.New()

func New(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var input types.StudentInput
		if !decodeInput(w, r, &input) || !validateInput(w, input) {
			return
		}

		student, err := store.CreateStudent(input.Name, input.Email, input.Age)
		if err != nil {
			slog.Error("failed to create student", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.Header().Set("Location", fmt.Sprintf("/api/students/%d", student.ID))
		writeJSON(w, http.StatusCreated, student)
	}
}

func GetStudentById(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := studentID(w, r)
		if !ok {
			return
		}

		student, err := store.GetStudentById(id)
		if err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			slog.Error("failed to get student", slog.Int64("id", id), slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, student)
	}
}

func GetList(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		students, err := store.GetStudents()
		if err != nil {
			slog.Error("failed to list students", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, students)
	}
}

func UpdateStudentById(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := studentID(w, r)
		if !ok {
			return
		}

		var input types.StudentInput
		if !decodeInput(w, r, &input) || !validateInput(w, input) {
			return
		}

		student, err := store.UpdateStudentById(id, input.Name, input.Email, input.Age)
		if err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			slog.Error("failed to update student", slog.Int64("id", id), slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		writeJSON(w, http.StatusOK, student)
	}
}

func DeleteStudentById(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, ok := studentID(w, r)
		if !ok {
			return
		}

		if err := store.DeleteStudentById(id); err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				writeError(w, http.StatusNotFound, err.Error())
				return
			}
			slog.Error("failed to delete student", slog.Int64("id", id), slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "internal server error")
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func studentID(w http.ResponseWriter, r *http.Request) (int64, bool) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil || id <= 0 {
		writeError(w, http.StatusBadRequest, "student id must be a positive integer")
		return 0, false
	}
	return id, true
}

func decodeInput(w http.ResponseWriter, r *http.Request, input *types.StudentInput) bool {
	r.Body = http.MaxBytesReader(w, r.Body, maxRequestBodyBytes)
	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields()

	if err := decoder.Decode(input); err != nil {
		if errors.Is(err, io.EOF) {
			writeError(w, http.StatusBadRequest, "request body must not be empty")
		} else {
			var maxBytesErr *http.MaxBytesError
			if errors.As(err, &maxBytesErr) {
				writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
			} else {
				writeError(w, http.StatusBadRequest, "invalid JSON request body")
			}
		}
		return false
	}

	var extra any
	if err := decoder.Decode(&extra); !errors.Is(err, io.EOF) {
		var maxBytesErr *http.MaxBytesError
		if errors.As(err, &maxBytesErr) {
			writeError(w, http.StatusRequestEntityTooLarge, "request body is too large")
		} else {
			writeError(w, http.StatusBadRequest, "request body must contain a single JSON object")
		}
		return false
	}

	input.Name = strings.TrimSpace(input.Name)
	input.Email = strings.TrimSpace(input.Email)
	return true
}

func validateInput(w http.ResponseWriter, input types.StudentInput) bool {
	if err := validate.Struct(input); err != nil {
		validationErrors, ok := err.(validator.ValidationErrors)
		if !ok {
			slog.Error("failed to validate student input", slog.String("error", err.Error()))
			writeError(w, http.StatusInternalServerError, "internal server error")
			return false
		}
		writeJSON(w, http.StatusBadRequest, response.ValidationError(validationErrors))
		return false
	}
	return true
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	if err := response.WriteJSON(w, status, body); err != nil {
		slog.Error("failed to write HTTP response", slog.String("error", err.Error()))
	}
}

func writeError(w http.ResponseWriter, status int, message string) {
	if err := response.WriteError(w, status, message); err != nil {
		slog.Error("failed to write HTTP error response", slog.String("error", err.Error()))
	}
}
