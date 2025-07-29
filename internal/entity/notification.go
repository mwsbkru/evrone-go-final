package entity

type Notification struct {
	UserEmail    string `json:"user_email"`
	Subject      string `json:"subject"`
	Body         string `json:"body"`
	CurrentRetry int
	Channel      string
}
