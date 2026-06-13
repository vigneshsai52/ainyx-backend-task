// Package models defines request/response structures used across layers.
package models

// CreateUserRequest is the payload for POST /users.
type CreateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	// DOB must be a valid date string in YYYY-MM-DD format.
	Dob string `json:"dob" validate:"required,datetime=2006-01-02"`
}

// UpdateUserRequest is the payload for PUT /users/:id.
type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=255"`
	Dob  string `json:"dob" validate:"required,datetime=2006-01-02"`
}

// UserResponse is returned for create / update operations (no age field).
type UserResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Dob  string `json:"dob"` // formatted as YYYY-MM-DD
}

// UserWithAgeResponse is returned for get / list operations (includes age).
type UserWithAgeResponse struct {
	ID   int32  `json:"id"`
	Name string `json:"name"`
	Dob  string `json:"dob"`
	Age  int    `json:"age"`
}

// PaginatedUsersResponse wraps the list response with metadata.
type PaginatedUsersResponse struct {
	Data       []UserWithAgeResponse `json:"data"`
	Total      int64                 `json:"total"`
	Page       int                   `json:"page"`
	PageSize   int                   `json:"page_size"`
	TotalPages int                   `json:"total_pages"`
}

// ErrorResponse is the standard error envelope.
type ErrorResponse struct {
	Error string `json:"error"`
}
