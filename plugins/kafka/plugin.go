// Package kafkaplugin implements the "kafka" protocol plugin for Wynbench.
//
// It produces a single message to a Kafka topic using github.com/segmentio/kafka-go.
// The plugin is intentionally simple (no consumer support, no SASL/TLS yet);
// those can be layered on without changing the core.Plugin interface.
package kafkaplugin

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	avro "github.com/iskorotkov/avro/v2"
	kafka "github.com/segmentio/kafka-go"

	"github.com/wynforge/wynbench-agent/core"
)

// writeTimeout bounds how long a single produce call may take.
const writeTimeout = 10 * time.Second

// Plugin is the Kafka protocol plugin.
type Plugin struct{}

// New returns a ready-to-use Kafka Plugin.
func New() *Plugin { return &Plugin{} }

// Name returns the protocol identifier used to look up this plugin.
func (p *Plugin) Name() string { return "kafka" }

// Configure accepts optional connection-level settings. Currently unused but
// required to satisfy the core.Plugin interface.
func (p *Plugin) Configure(_ map[string]any) error { return nil }

// Execute performs Kafka operations for the plugin.
//
// Supported operations:
//
//	"produce"      – publish a single message (default)
//	"list_topics"  – return available topics for the configured brokers
//	"read_messages" – read live messages for a topic and partition
//
// Common params:
//
//	"brokers" (string) – comma-separated broker list, usually from connection config
//	"operation" (string) – one of produce, list_topics, read_messages
//
// Produce params:
//
//	"topic"   (string) – destination topic
//	"value"   (any) – message payload
//	"key"     (string) – optional message key
//	"headers" (map[string]any) – optional headers
func (p *Plugin) Execute(action core.Action) (core.Result, error) {
	operation := "produce"
	if op, ok := stringParam(action.Params, "operation"); ok {
		operation = strings.ToLower(strings.TrimSpace(op))
	}

	switch operation {
	case "list_topics":
		return p.listTopics(action)
	case "read_messages":
		return p.readMessages(action)
	case "produce", "send", "publish":
		return p.produceMessage(action)
	default:
		return core.Result{Success: false, Error: fmt.Sprintf("unknown kafka operation %q", operation)}, nil
	}
}

func (p *Plugin) listTopics(action core.Action) (core.Result, error) {
	brokers, ok := stringParam(action.Params, "brokers")
	if !ok {
		return core.Result{Success: false, Error: "missing param: brokers"}, nil
	}
	addrs := splitBrokers(brokers)
	if len(addrs) == 0 {
		return core.Result{Success: false, Error: "param brokers must contain at least one broker address"}, nil
	}

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	conn, err := kafka.DialContext(ctx, "tcp", addrs[0])
	if err != nil {
		return core.Result{Success: false, Error: fmt.Sprintf("failed to connect to Kafka broker: %v", err)}, nil
	}
	defer conn.Close()

	partitions, err := conn.ReadPartitions()
	if err != nil {
		return core.Result{Success: false, Error: fmt.Sprintf("failed to read Kafka metadata: %v", err)}, nil
	}

	topicsMap := make(map[string]struct{})
	for _, p := range partitions {
		topicsMap[p.Topic] = struct{}{}
	}

	topics := make([]string, 0, len(topicsMap))
	for topic := range topicsMap {
		topics = append(topics, topic)
	}
	sort.Strings(topics)

	return core.Result{Success: true, Data: map[string]any{"topics": topics}}, nil
}

func (p *Plugin) readMessages(action core.Action) (core.Result, error) {
	brokers, ok := stringParam(action.Params, "brokers")
	if !ok {
		return core.Result{Success: false, Error: "missing param: brokers"}, nil
	}
	addrs := splitBrokers(brokers)
	if len(addrs) == 0 {
		return core.Result{Success: false, Error: "param brokers must contain at least one broker address"}, nil
	}

	topic, ok := stringParam(action.Params, "topic")
	if !ok {
		return core.Result{Success: false, Error: "missing param: topic"}, nil
	}

	partition := 0
	if param, ok := intParam(action.Params, "partition"); ok {
		partition = param
	}

	count := 20
	if param, ok := intParam(action.Params, "count"); ok {
		count = param
	}

	reader := kafka.NewReader(kafka.ReaderConfig{
		Brokers:   addrs,
		Topic:     topic,
		Partition: partition,
		MinBytes:  1,
		MaxBytes:  10e6,
	})
	defer reader.Close()

	messages := make([]map[string]any, 0, count)
	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	for i := 0; i < count; i++ {
		msg, err := reader.ReadMessage(ctx)
		if err != nil {
			if err == context.Canceled || err == context.DeadlineExceeded {
				break
			}
			return core.Result{Success: false, Error: fmt.Sprintf("failed to read messages: %v", err)}, nil
		}
		messages = append(messages, map[string]any{
			"offset":    msg.Offset,
			"partition": msg.Partition,
			"key":       string(msg.Key),
			"value":     string(msg.Value),
			"headers":   headersToMap(msg.Headers),
			"time":      msg.Time,
		})
	}

	return core.Result{Success: true, Data: map[string]any{"messages": messages}}, nil
}

func (p *Plugin) produceMessage(action core.Action) (core.Result, error) {
	brokers, ok := stringParam(action.Params, "brokers")
	if !ok {
		return core.Result{Success: false, Error: "missing param: brokers"}, nil
	}
	addrs := splitBrokers(brokers)
	if len(addrs) == 0 {
		return core.Result{Success: false, Error: "param brokers must contain at least one broker address"}, nil
	}

	topic, ok := stringParam(action.Params, "topic")
	if !ok {
		return core.Result{Success: false, Error: "missing param: topic"}, nil
	}

	valueBytes, err := marshalActionValue(action.Params["value"])
	if err != nil {
		return core.Result{Success: false, Error: err.Error()}, nil
	}

	if schema, ok := stringParam(action.Params, "avro_schema"); ok {
		valueBytes, err = encodeAvroValue(schema, action.Params["value"])
		if err != nil {
			return core.Result{Success: false, Error: err.Error()}, nil
		}
	}

	msg := kafka.Message{Value: valueBytes}
	if key, ok := stringParam(action.Params, "key"); ok {
		msg.Key = []byte(key)
	}
	if headers, ok := action.Params["headers"].(map[string]any); ok {
		for k, v := range headers {
			msg.Headers = append(msg.Headers, kafka.Header{Key: k, Value: []byte(fmt.Sprint(v))})
		}
	}

	writer := &kafka.Writer{
		Addr:                   kafka.TCP(addrs...),
		Topic:                  topic,
		Balancer:               &kafka.LeastBytes{},
		AllowAutoTopicCreation: true,
	}
	defer writer.Close()

	ctx, cancel := context.WithTimeout(context.Background(), writeTimeout)
	defer cancel()

	if err := writer.WriteMessages(ctx, msg); err != nil {
		return core.Result{Success: false, Error: fmt.Sprintf("kafka produce failed: %v", err)}, nil
	}

	return core.Result{
		Success: true,
		Data: map[string]any{
			"topic":   topic,
			"brokers": addrs,
			"key":     string(msg.Key),
		},
	}, nil
}

// stringParam reads a non-empty string parameter from params.
func stringParam(params map[string]any, key string) (string, bool) {
	v, ok := params[key].(string)
	if !ok || strings.TrimSpace(v) == "" {
		return "", false
	}
	return v, true
}

func intParam(params map[string]any, key string) (int, bool) {
	switch value := params[key].(type) {
	case int:
		return value, true
	case float64:
		return int(value), true
	case string:
		var parsed int
		if _, err := fmt.Sscanf(strings.TrimSpace(value), "%d", &parsed); err == nil {
			return parsed, true
		}
	}
	return 0, false
}

// splitBrokers parses a comma-separated broker list, trimming whitespace and
// dropping empty entries.
func marshalActionValue(raw any) ([]byte, error) {
	switch typed := raw.(type) {
	case string:
		trimmed := strings.TrimSpace(typed)
		if json.Valid([]byte(trimmed)) {
			return []byte(trimmed), nil
		}
		return []byte(typed), nil
	case []byte:
		if json.Valid(typed) {
			return typed, nil
		}
		return typed, nil
	case map[string]any, []any:
		return json.Marshal(typed)
	default:
		return json.Marshal(typed)
	}
}

func encodeAvroValue(schemaText string, raw any) ([]byte, error) {
	schema, err := avro.Parse(schemaText)
	if err != nil {
		return nil, fmt.Errorf("invalid avro_schema: %w", err)
	}

	valueBytes, err := marshalActionValue(raw)
	if err != nil {
		return nil, err
	}

	var decoded any
	if err := json.Unmarshal(valueBytes, &decoded); err != nil {
		return nil, fmt.Errorf("avro value must be valid JSON: %w", err)
	}

	return avro.Marshal(schema, decoded)
}

func splitBrokers(raw string) []string {
	parts := strings.Split(raw, ",")
	addrs := make([]string, 0, len(parts))
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if trimmed != "" {
			addrs = append(addrs, trimmed)
		}
	}
	return addrs
}

func headersToMap(headers []kafka.Header) map[string]string {
	result := make(map[string]string, len(headers))
	for _, header := range headers {
		result[header.Key] = string(header.Value)
	}
	return result
}
