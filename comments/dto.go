package comments

type CreateRequest struct {
	PostID  uint   `json:"post_id" validate:"required"`
	Content string `json:"content" validate:"required"`
}

type UpdateRequest struct {
	Content *string `json:"content,omitempty" validate:"omitempty,min=1"`
}

type Response struct {
	ID        uint   `json:"id"`
	PostID    uint   `json:"post_id"`
	Content   string `json:"content"`
	UserID    uint   `json:"user_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type ListResponse struct {
	Comments []Response `json:"comments"`
	Total    int64      `json:"total"`
	Limit    int        `json:"limit"`
	Offset   int        `json:"offset"`
}
