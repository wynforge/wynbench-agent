package api

import (
	"net/http"
	"strconv"

	"github.com/wynforge/wynbench-agent/core"
)

// listKafkaTopics handles GET /kafka/topics?brokers=... and returns available topics.
func (s *Server) listKafkaTopics(w http.ResponseWriter, r *http.Request) {
	brokers := r.URL.Query().Get("brokers")
	if brokers == "" {
		writeError(w, http.StatusBadRequest, "brokers are required")
		return
	}

	action := core.Action{
		Plugin: "kafka",
		Params: map[string]any{
			"brokers":   brokers,
			"operation": "list_topics",
		},
	}

	result, err := s.engine.ExecuteAction(action)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, result)
		return
	}
	if !result.Success {
		writeError(w, http.StatusBadRequest, result.Error)
		return
	}

	writeJSON(w, http.StatusOK, result.Data)
}

// listKafkaTopicMessages handles GET /kafka/messages?brokers=...&topic=...&partition=0&count=20
func (s *Server) listKafkaTopicMessages(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	brokers := query.Get("brokers")
	topic := query.Get("topic")
	partitionStr := query.Get("partition")
	countStr := query.Get("count")

	if brokers == "" || topic == "" {
		writeError(w, http.StatusBadRequest, "brokers and topic are required")
		return
	}

	partition := 0
	if partitionStr != "" {
		parsed, err := strconv.Atoi(partitionStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "partition must be an integer")
			return
		}
		partition = parsed
	}

	count := 20
	if countStr != "" {
		parsed, err := strconv.Atoi(countStr)
		if err != nil {
			writeError(w, http.StatusBadRequest, "count must be an integer")
			return
		}
		count = parsed
	}

	action := core.Action{
		Plugin: "kafka",
		Params: map[string]any{
			"brokers":   brokers,
			"operation": "read_messages",
			"topic":     topic,
			"partition": partition,
			"count":     count,
		},
	}

	result, err := s.engine.ExecuteAction(action)
	if err != nil {
		writeJSON(w, http.StatusUnprocessableEntity, result)
		return
	}
	if !result.Success {
		writeError(w, http.StatusBadRequest, result.Error)
		return
	}

	writeJSON(w, http.StatusOK, result.Data)
}
