package models

type EmailMessage struct {
	Email       string `json:"to"`
	MessageText string `json:"link"`
	Subject     string `json:"subject"`
}
