package domain

import "time"

type Request struct {
	ID      string              `json:"id"`
	Name    string              `json:"name"`
	Method  string              `json:"method"`
	URL     string              `json:"url"`
	Headers map[string][]string `json:"headers"`
	Body    string              `json:"body"`
}

type Response struct {
	Status   int                 `json:"status"`
	Headers  map[string][]string `json:"headers"`
	Body     []byte              `json:"body"`
	Duration time.Duration       `json:"duration"`
}

type Collection struct {
	Variables map[string]string `json:"variables"`
	Requests  []Request         `json:"requests"`
}
