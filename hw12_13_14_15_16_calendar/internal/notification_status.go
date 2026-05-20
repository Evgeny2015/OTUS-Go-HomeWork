package internal

import (
	"encoding/json"
	"time"
)

// NotificationStatus represents the status of a processed notification
type NotificationStatus struct {
	EventID    string    `json:"eventId"`
	Status     string    `json:"status"` // "processed", "failed"
	Timestamp  time.Time `json:"timestamp"`
	Error      string    `json:"error,omitempty"`
	UserID     string    `json:"userId"`
	EventTitle string    `json:"eventTitle,omitempty"`
	EventDate  time.Time `json:"eventDate,omitempty"`
}

// NewNotificationStatus creates a new NotificationStatus
func NewNotificationStatus(eventID, status, userID string, timestamp time.Time) *NotificationStatus {
	return &NotificationStatus{
		EventID:   eventID,
		Status:    status,
		UserID:    userID,
		Timestamp: timestamp,
	}
}

// WithError sets the error message
func (ns *NotificationStatus) WithError(err string) *NotificationStatus {
	ns.Error = err
	return ns
}

// WithEventDetails sets event title and date
func (ns *NotificationStatus) WithEventDetails(title string, date time.Time) *NotificationStatus {
	ns.EventTitle = title
	ns.EventDate = date
	return ns
}

// ToJSON serializes notification status to JSON
func (ns *NotificationStatus) ToJSON() ([]byte, error) {
	return json.Marshal(ns)
}

// FromJSON deserializes notification status from JSON
func NotificationStatusFromJSON(data []byte) (*NotificationStatus, error) {
	var status NotificationStatus
	err := json.Unmarshal(data, &status)
	return &status, err
}
