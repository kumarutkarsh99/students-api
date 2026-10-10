package student

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/kumarutkarsh99/students-api/internal/storage"
	"github.com/kumarutkarsh99/students-api/internal/types"
	"github.com/kumarutkarsh99/students-api/internal/utils/response"
)

type updateStudentRequest struct {
	Name  string `json:"name" validate:"required"`
	Age   int    `json:"age" validate:"gt=0"`
	Email string `json:"email" validate:"required"`
}

func New(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var student types.Student
		err := json.NewDecoder(r.Body).Decode(&student)
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// request validation
		if err := validator.New().Struct(student); err != nil {
			validateErrs := err.(validator.ValidationErrors)
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		slog.Info("creating a student")

		lastId, err := storage.CreateStudent(
			student.Name,
			student.Email,
			student.Age,
		)

		if err != nil {
			slog.Error("failed to create student", slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, err)
			return
		}

		slog.Info("student created successfully", slog.String("studentId", fmt.Sprint(lastId)))
		response.WriteJson(w, http.StatusCreated, map[string]int64{"id": lastId})
	}
}

func GetStudentById(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("getting a student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		student, err := storage.GetStudentById(intId)
		if err != nil {
			slog.Error("error getting user", slog.String("id", id))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		response.WriteJson(w, http.StatusOK, student)
	}
}

func GetList(storage storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slog.Info("getting all students")
		students, err := storage.GetStudents()
		if err != nil {
			response.WriteJson(w, http.StatusInternalServerError, err)
		}
		response.WriteJson(w, http.StatusOK, students)
	}
}

func DeleteStudentById(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("deleting student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if intId <= 0 {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("student id must be a positive integer")))
			return
		}

		err = store.DeleteStudentById(intId)
		if err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return
			}

			slog.Error("failed to delete student", slog.String("id", id), slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}

func UpdateStudentById(store storage.Storage) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		slog.Info("updating student", slog.String("id", id))

		intId, err := strconv.ParseInt(id, 10, 64)
		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		if intId <= 0 {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("student id must be a positive integer")))
			return
		}

		var student updateStudentRequest
		err = json.NewDecoder(r.Body).Decode(&student)
		if errors.Is(err, io.EOF) {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(fmt.Errorf("empty body")))
			return
		}

		if err != nil {
			response.WriteJson(w, http.StatusBadRequest, response.GeneralError(err))
			return
		}

		// request validation
		if err := validator.New().Struct(student); err != nil {
			validateErrs, ok := err.(validator.ValidationErrors)
			if !ok {
				response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
				return
			}
			response.WriteJson(w, http.StatusBadRequest, response.ValidationError(validateErrs))
			return
		}

		err = store.UpdateStudentById(intId, student.Name, student.Email, student.Age)
		if err != nil {
			if errors.Is(err, storage.ErrStudentNotFound) {
				response.WriteJson(w, http.StatusNotFound, response.GeneralError(err))
				return
			}

			slog.Error("failed to update student", slog.String("id", id), slog.String("error", err.Error()))
			response.WriteJson(w, http.StatusInternalServerError, response.GeneralError(err))
			return
		}

		w.WriteHeader(http.StatusNoContent)
	}
}
