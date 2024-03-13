package models

type Finding struct {
	Id     string `json:"id"`
	Type   string `json:"type"`
	Reason string `json:"reason"`
}
