package eventstore

import (
	"encoding/json"
	"fmt"
	"time"

	"dag-observatory/dag-core/internal/domain/dagruntime/events"
	"dag-observatory/dag-core/internal/domain/dagruntime/state"
)

type kafkaEventMessage struct {
	EventID   string                  `json:"event_id"`
	EventTime string                  `json:"event_time"`
	Partition string                  `json:"partition"`
	Type      string                  `json:"type"`
	Payload   []events.PayloadEnvelope `json:"payload,omitempty"`
}

func EncodeEvent(event events.Event) ([]byte, []byte, error) {
	var payload []events.PayloadEnvelope
	switch value := event.Payload.(type) {
	case nil:
		payload = nil
	case []events.PayloadEnvelope:
		payload = value
	default:
		return nil, nil, fmt.Errorf("eventstore: unsupported payload type %T", event.Payload)
	}
	msg := kafkaEventMessage{
		EventID:   event.EventID,
		EventTime: event.EventTime.Format("2006-01-02T15:04:05.999999999Z07:00"),
		Partition: string(event.Partition),
		Type:      event.Type,
		Payload:   payload,
	}
	value, err := json.Marshal(msg)
	if err != nil {
		return nil, nil, err
	}
	key := []byte(msg.Partition)
	return key, value, nil
}

func DecodeEvent(key []byte, value []byte) (events.Event, error) {
	var msg kafkaEventMessage
	if err := json.Unmarshal(value, &msg); err != nil {
		return events.Event{}, err
	}
	if msg.Partition == "" && len(key) > 0 {
		msg.Partition = string(key)
	}
	eventTime := time.Time{}
	if msg.EventTime != "" {
		parsed, err := parseEventTime(msg.EventTime)
		if err != nil {
			return events.Event{}, err
		}
		eventTime = parsed
	}
	return events.Event{
		EventID:   msg.EventID,
		EventTime: eventTime,
		Partition: state.Partition(msg.Partition),
		Type:      msg.Type,
		Payload:   msg.Payload,
	}, nil
}

func parseEventTime(raw string) (time.Time, error) {
	t, err := time.Parse("2006-01-02T15:04:05.999999999Z07:00", raw)
	if err == nil {
		return t, nil
	}
	return time.Parse(time.RFC3339Nano, raw)
}
