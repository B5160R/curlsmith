package tui

import (
	"fmt"
	"sort"
	"strings"
)

func parseHeaders(raw string) (map[string][]string, error) {
	headers := make(map[string][]string)

	for lineNumber, line := range strings.Split(raw, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		parts := strings.SplitN(line, ":", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("header line %d must look like Key: Value", lineNumber+1)
		}

		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		if key == "" {
			return nil, fmt.Errorf("header line %d has an empty key", lineNumber+1)
		}

		headers[key] = append(headers[key], value)
	}

	return headers, nil
}

func renderHeadersInput(headers map[string][]string) string {
	if len(headers) == 0 {
		return ""
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0)
	for _, key := range keys {
		for _, value := range headers[key] {
			lines = append(lines, fmt.Sprintf("%s: %s", key, value))
		}
	}

	return strings.Join(lines, "\n")
}

func formatHeaders(headers map[string][]string) string {
	if len(headers) == 0 {
		return "(none)"
	}

	keys := make([]string, 0, len(headers))
	for key := range headers {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(headers))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("%s: %s", key, strings.Join(headers[key], ", ")))
	}

	return strings.Join(lines, "\n")
}

func formatVariables(vars map[string]string) []string {
	keys := make([]string, 0, len(vars))
	for key := range vars {
		keys = append(keys, key)
	}
	sort.Strings(keys)

	lines := make([]string, 0, len(keys))
	for _, key := range keys {
		lines = append(lines, fmt.Sprintf("- %s = %s", key, vars[key]))
	}
	return lines
}

func findMethodIndex(method string) int {
	method = strings.ToUpper(strings.TrimSpace(method))
	for index, candidate := range methods {
		if candidate == method {
			return index
		}
	}
	return 0
}
