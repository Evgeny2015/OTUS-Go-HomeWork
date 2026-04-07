package storage

import "time"

type Event struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	DateTime     time.Time     `json:"date_time"`
	Duration     time.Duration `json:"duration"`
	Description  string        `json:"description,omitempty"`
	UserID       string        `json:"user_id"`
	NotifyBefore time.Duration `json:"notify_before,omitempty"`
}
