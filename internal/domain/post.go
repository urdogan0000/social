package domain

type PostID uint

type Post struct {
	ID      PostID
	Title   string
	Content string
	UserID  UserID
	Tags    []string
}

// Validate validates post data
func (p *Post) Validate() error {
	if len(p.Title) == 0 || len(p.Title) > 255 {
		return ErrInvalidTitle
	}
	if len(p.Content) == 0 {
		return ErrInvalidContent
	}
	return nil
}

// CanBeEditedBy checks if the post can be edited by the given user
func (p *Post) CanBeEditedBy(userID UserID) bool {
	return p.UserID == userID
}

// CanBeDeletedBy checks if the post can be deleted by the given user
func (p *Post) CanBeDeletedBy(userID UserID) bool {
	return p.UserID == userID
}

// UpdateTitle updates the title if valid
func (p *Post) UpdateTitle(newTitle string) error {
	if len(newTitle) == 0 || len(newTitle) > 255 {
		return ErrInvalidTitle
	}
	p.Title = newTitle
	return nil
}

// UpdateContent updates the content if valid
func (p *Post) UpdateContent(newContent string) error {
	if len(newContent) == 0 {
		return ErrInvalidContent
	}
	p.Content = newContent
	return nil
}

// UpdateTags updates the tags
func (p *Post) UpdateTags(newTags []string) {
	p.Tags = newTags
}

// HasTag checks if post has a specific tag
func (p *Post) HasTag(tag string) bool {
	for _, t := range p.Tags {
		if t == tag {
			return true
		}
	}
	return false
}

