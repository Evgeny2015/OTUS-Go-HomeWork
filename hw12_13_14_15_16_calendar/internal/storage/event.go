package storage

import "time"

type Event struct {
	ID           string        `json:"id"`
	Title        string        `json:"title"`
	DateTime     time.Time     `json:"dateTime"`
	Duration     time.Duration `json:"duration"`
	Description  string        `json:"description,omitempty"`
	UserID       string        `json:"userId"`
	NotifyBefore time.Duration `json:"notifyBefore,omitempty"`
}
