package entity

type Team struct {
	Name    string `json:"team_name"`
	Members []User `json:"members"`
}
