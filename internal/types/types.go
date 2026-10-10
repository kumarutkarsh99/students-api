package types

type Student struct {
	ID    int64  `json:"id"`
	Name  string `json:"name"`
	Age   int    `json:"age"`
	Email string `json:"email"`
}

type StudentInput struct {
	Name  string `json:"name" validate:"required,max=100"`
	Age   int    `json:"age" validate:"gt=0"`
	Email string `json:"email" validate:"required,email,max=254"`
}
