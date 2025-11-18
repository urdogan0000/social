package users

// CreateRequest represents the request to create a user
type CreateRequest struct {
	Username string `json:"username" validate:"required,min=3,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// UpdateRequest represents the request to update a user
type UpdateRequest struct {
	Username *string `json:"username,omitempty" validate:"omitempty,min=3,max=100"`
	Email    *string `json:"email,omitempty" validate:"omitempty,email"`
	Password *string `json:"password,omitempty" validate:"omitempty,min=6"`
}

// Response represents a user in API responses
type Response struct {
	ID        uint   `json:"id"`
	Username  string  `json:"username"`
	Email     string  `json:"email"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ListResponse represents a paginated list of users
type ListResponse struct {
	Users []Response `json:"users"`
	Total int64      `json:"total"`
	Limit int        `json:"limit"`
	Offset int       `json:"offset"`
}

