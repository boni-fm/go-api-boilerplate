package models

// ProfileResponse is the response DTO for the User Profile endpoint.
type ProfileResponse struct {
	UserName    string `json:"user_name"    db:"user_name"`
	DisplayName string `json:"display_name" db:"display_name"`
	Email       string `json:"email"        db:"email"`
}

// UpdateProfileRequest is the request DTO for creating or updating a user profile.
type UpdateProfileRequest struct {
	DisplayName string `json:"display_name" example:"Alice Smith"`
	Email       string `json:"email"        example:"alice@example.com"`
}
