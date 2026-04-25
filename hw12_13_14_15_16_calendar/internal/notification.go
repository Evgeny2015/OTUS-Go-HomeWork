package internal

import (
	"encoding/json"
	"time"
)

// Notification is a temporary entity, not stored in DB, placed in queue for sender
type Notification struct {
	EventID    string    `json:"eventId"`
	EventTitle string    `json:"eventTitle"`
	EventDate  time.Time `json:"eventDate"`
	UserID     string    `json:"userId"`
}

// NewNotification creates a new Notification from event data
func NewNotification(eventID, title string, eventDate time.Time, userID string) *Notification {
	return &Notification{
		EventID:    eventID,
		EventTitle: title,
		EventDate:  eventDate,
		UserID:     userID,
	}
}

// ToJSON serializes notification to JSON
func (n *Notification) ToJSON() ([]byte, error) {
	return json.Marshal(n)
}

// FromJSON deserializes notification from JSON
func FromJSON(data []byte) (*Notification, error) {
	var notification Notification
	err := json.Unmarshal(data, &notification)
	return &notification, err
}
