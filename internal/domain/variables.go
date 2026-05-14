package domain

import "regexp"

var variablePattern = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)

// Resolve replaces {{variable}} placeholders in input with values from vars.
func Resolve(input string, vars map[string]string) string {
	if len(vars) == 0 || input == "" {
		return input
	}

	return variablePattern.ReplaceAllStringFunc(input, func(match string) string {
		parts := variablePattern.FindStringSubmatch(match)
		if len(parts) != 2 {
			return match
		}

		value, ok := vars[parts[1]]
		if !ok {
			return match
		}

		return value
	})
}
