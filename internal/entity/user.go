package entity

type User struct {
	ID       string `json:"user_id"`
	Name     string `json:"username"`
	Team     string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}
