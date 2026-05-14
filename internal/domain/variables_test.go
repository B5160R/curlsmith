package domain

import "testing"

func TestResolveReplacesKnownVariables(t *testing.T) {
	input := "{{base_url}}/users/{{user_id}}"
	vars := map[string]string{
		"base_url": "https://api.example.com",
		"user_id":  "42",
	}

	got := Resolve(input, vars)
	want := "https://api.example.com/users/42"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolveLeavesUnknownVariables(t *testing.T) {
	got := Resolve("{{base_url}}/{{missing}}", map[string]string{"base_url": "https://api.example.com"})
	want := "https://api.example.com/{{missing}}"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}

func TestResolveEmptyInputReturnsEmpty(t *testing.T) {
	got := Resolve("", map[string]string{"key": "value"})
	if got != "" {
		t.Fatalf("expected empty string, got %q", got)
	}
}

func TestResolveNilVarsReturnsInput(t *testing.T) {
	got := Resolve("{{key}}", nil)
	if got != "{{key}}" {
		t.Fatalf("expected unchanged input, got %q", got)
	}
}

func TestResolveHandlesWhitespaceInBraces(t *testing.T) {
	got := Resolve("{{ base_url }}", map[string]string{"base_url": "http://localhost"})
	want := "http://localhost"
	if got != want {
		t.Fatalf("expected %q, got %q", want, got)
	}
}
