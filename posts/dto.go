package posts

type CreateRequest struct {
	Title   string   `json:"title" validate:"required,min=1,max=255"`
	Content string   `json:"content" validate:"required,min=1"`
	Tags    []string `json:"tags,omitempty"`
}

type UpdateRequest struct {
	Title   *string   `json:"title,omitempty" validate:"omitempty,min=1,max=255"`
	Content *string   `json:"content,omitempty" validate:"omitempty,min=1"`
	Tags    *[]string `json:"tags,omitempty"`
}

type Response struct {
	ID        uint     `json:"id"`
	Title     string   `json:"title"`
	Content   string   `json:"content"`
	UserID    uint     `json:"user_id"`
	Tags      []string `json:"tags"`
	CreatedAt string   `json:"created_at"`
	UpdatedAt string   `json:"updated_at"`
}

type ListResponse struct {
	Posts []Response `json:"posts"`
	Total int64      `json:"total"`
	Limit int        `json:"limit"`
	Offset int       `json:"offset"`
}

