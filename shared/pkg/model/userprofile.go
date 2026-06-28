package model

type UserProfile struct {
	ID        string   `json:"id"`
	Email     string   `json:"email"`
	Status    string   `json:"status"`
	Languages []string `json:"languages"`
}
