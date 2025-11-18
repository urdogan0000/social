package users

type CreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

type UpdateRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=100"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=6"`
}

type Response struct {
	ID        uint   `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListResponse struct {
	Users []Response `json:"users"`
	Total int64      `json:"total"`
	Limit int        `json:"limit"`
	Offset int       `json:"offset"`
}

