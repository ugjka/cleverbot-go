package cleverbot

import "testing"

func TestGetAPIID(t *testing.T) {
	_, err := getAPIID(protocol + host)
	if err != nil {
		t.Error("Expected no error, got", err)
	}
}

func TestSession(t *testing.T) {
	session, err := New()
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}
	answer, err := session.Ask("Hello World")
	if err != nil {
		t.Error("Expected no error, got", err)
		return
	}
	if answer == "" {
		t.Error("Expected response, got empty string")
		return
	}
	if answer == "Error: ask again!" {
		t.Error("Expected response, got Error")
	}
}
