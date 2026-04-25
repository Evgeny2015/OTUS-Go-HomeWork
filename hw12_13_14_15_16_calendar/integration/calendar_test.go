package integration

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
	"time"

	amqp "github.com/rabbitmq/amqp091-go"
)

const (
	calendarHost = "http://calendar:8080"
)

func TestHealthEndpoint(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	resp, err := client.Get(calendarHost + "/health")
	if err != nil {
		t.Fatalf("Failed to reach health endpoint: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestCreateEvent(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	event := map[string]interface{}{
		"id":           "test-event-1",
		"title":        "Test Event",
		"dateTime":     time.Now().Add(2 * time.Hour).Format(time.RFC3339),
		"duration":     int64(time.Hour),
		"description":  "Integration test event",
		"userId":       "user-1",
		"notifyBefore": int64(30 * time.Minute),
	}

	body, _ := json.Marshal(event)
	resp, err := client.Post(calendarHost+"/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Errorf("Expected status 201, got %d", resp.StatusCode)
	}

	var result map[string]string
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Errorf("Failed to decode response: %v", err)
	}
	if result["id"] != event["id"] {
		t.Errorf("Expected event ID %s, got %s", event["id"], result["id"])
	}
}

func TestCreateDuplicateEvent(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	event := map[string]interface{}{
		"id":           "duplicate-event",
		"title":        "Duplicate Event",
		"dateTime":     time.Now().Add(3 * time.Hour).Format(time.RFC3339),
		"duration":     int64(time.Hour),
		"userId":       "user-1",
		"notifyBefore": int64(0),
	}

	body, _ := json.Marshal(event)
	resp, err := client.Post(calendarHost+"/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create first event: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("First creation failed with status %d", resp.StatusCode)
	}

	// Try to create duplicate
	resp2, err := client.Post(calendarHost+"/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to send duplicate request: %v", err)
	}
	defer resp2.Body.Close()

	// Expect error (should be 409 or 500? depends on implementation)
	if resp2.StatusCode == http.StatusCreated {
		t.Error("Duplicate event creation should not succeed")
	}
}

func TestListEventsForDay(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	// Create an event for today
	now := time.Now()
	event := map[string]interface{}{
		"id":           "list-event-day",
		"title":        "Day Event",
		"dateTime":     now.Add(1 * time.Hour).Format(time.RFC3339),
		"duration":     int64(30 * time.Minute),
		"userId":       "user-2",
		"notifyBefore": int64(0),
	}
	body, _ := json.Marshal(event)
	resp, err := client.Post(calendarHost+"/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create event for day list: %v", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Creation failed with status %d", resp.StatusCode)
	}

	// List events for day
	dateParam := now.Format("2006-01-02")
	url := fmt.Sprintf("%s/events/day?date=%sT00:00:00Z", calendarHost, dateParam)
	resp, err = client.Get(url)
	if err != nil {
		t.Fatalf("Failed to list events for day: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	var events []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		t.Errorf("Failed to decode events list: %v", err)
	}
	if len(events) == 0 {
		t.Error("Expected at least one event for the day")
	}
}

func TestListEventsForWeek(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	now := time.Now()
	url := fmt.Sprintf("%s/events/week?date=%sT00:00:00Z", calendarHost, now.Format("2006-01-02"))
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to list events for week: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	// Just ensure response is valid JSON
	var events []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		t.Errorf("Failed to decode week events list: %v", err)
	}
}

func TestListEventsForMonth(t *testing.T) {
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	now := time.Now()
	url := fmt.Sprintf("%s/events/month?date=%sT00:00:00Z", calendarHost, now.Format("2006-01-02"))
	resp, err := client.Get(url)
	if err != nil {
		t.Fatalf("Failed to list events for month: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	var events []map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&events); err != nil {
		t.Errorf("Failed to decode month events list: %v", err)
	}
}

// TestNotificationStatus verifies that notification status is sent when an event with notification time is created.
// This test creates an event with a notification time in the past (so it's due immediately),
// then checks that a notification status appears in the RabbitMQ status queue.
func TestNotificationStatus(t *testing.T) {
	// RabbitMQ connection parameters (should match docker-compose.integration.yaml)
	const (
		rabbitMQURI      = "amqp://rabbit:password@rabbitmq:5672/"
		exchangeName     = "calendar_exchange"
		statusQueueName  = "notification_status"
		statusRoutingKey = "notification_status"
	)

	// Connect to RabbitMQ
	conn, err := amqp.Dial(rabbitMQURI)
	if err != nil {
		t.Fatalf("Failed to connect to RabbitMQ: %v", err)
	}
	defer conn.Close()

	ch, err := conn.Channel()
	if err != nil {
		t.Fatalf("Failed to open channel: %v", err)
	}
	defer ch.Close()

	// Ensure status queue exists (declare it)
	_, err = ch.QueueDeclare(
		statusQueueName, // name
		true,            // durable
		false,           // delete when unused
		false,           // exclusive
		false,           // no-wait
		nil,             // arguments
	)
	if err != nil {
		t.Fatalf("Failed to declare status queue: %v", err)
	}

	// Bind queue to exchange with routing key
	err = ch.QueueBind(
		statusQueueName,  // queue name
		statusRoutingKey, // routing key
		exchangeName,     // exchange
		false,            // no-wait
		nil,              // arguments
	)
	if err != nil {
		t.Fatalf("Failed to bind status queue: %v", err)
	}

	// Purge existing messages from status queue to start fresh
	_, err = ch.QueuePurge(statusQueueName, false)
	if err != nil {
		t.Fatalf("Failed to purge status queue: %v", err)
	}

	// Start consuming from status queue
	msgs, err := ch.Consume(
		statusQueueName, // queue
		"test-consumer", // consumer
		false,           // auto-ack
		false,           // exclusive
		false,           // no-local
		false,           // no-wait
		nil,             // args
	)
	if err != nil {
		t.Fatalf("Failed to consume from status queue: %v", err)
	}

	// Create an event with notification time in the near past
	client := &http.Client{
		Timeout: 5 * time.Second,
	}

	eventID := fmt.Sprintf("notification-test-%d", time.Now().UnixNano())
	eventTime := time.Now().Add(1 * time.Minute) // Event in 1 minute
	notifyBefore := int64(65 * time.Second)      // Notification time = eventTime - 65s = 5 seconds ago

	event := map[string]interface{}{
		"id":           eventID,
		"title":        "Notification Test Event",
		"dateTime":     eventTime.Format(time.RFC3339),
		"duration":     int64(30 * time.Minute),
		"description":  "Event for notification status test",
		"userId":       "notification-test-user",
		"notifyBefore": notifyBefore,
	}

	body, _ := json.Marshal(event)
	resp, err := client.Post(calendarHost+"/events", "application/json", bytes.NewReader(body))
	if err != nil {
		t.Fatalf("Failed to create event: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		t.Fatalf("Expected status 201, got %d. Response body may contain error details", resp.StatusCode)
	}

	// Wait for notification status (up to 2 minutes)
	timeout := time.After(2 * time.Minute)
	ticker := time.NewTicker(1 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			t.Fatal("Timeout waiting for notification status")
		case <-ticker.C:
			// Check if there's a message available (non-blocking)
			select {
			case msg := <-msgs:
				// Parse notification status
				var status map[string]interface{}
				if err := json.Unmarshal(msg.Body, &status); err != nil {
					t.Errorf("Failed to unmarshal status message: %v", err)
					msg.Nack(false, false) // reject
					continue
				}

				// Verify status
				if status["eventId"] != eventID {
					t.Logf("Received status for different event: %s, expected %s", status["eventId"], eventID)
					msg.Ack(false)
					continue // keep waiting
				}

				if status["status"] != "processed" {
					t.Errorf("Expected status 'processed', got %s", status["status"])
				}

				t.Logf("Successfully received notification status for event %s", eventID)
				msg.Ack(false)
				return // test passed
			default:
				// No message yet, continue waiting
			}
		}
	}
}
