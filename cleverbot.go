// Package cleverbot is a wrapper for cleverbot.com api.
// To get a new session call New("YOURAPIKEY") and to ask call Session.Ask(question).
// Get the official API Key here http://www.cleverbot.com/api/ .
package cleverbot

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

var (
	host     = "www.cleverbot.com"
	protocol = "http://"
	resource = "/getreply?"
	apiURL   = protocol + host + resource
)

//Session is a cleverbot session.
type Session struct {
	client  *http.Client
	values  *url.Values
	decoder map[string]interface{}
}

//New creates a new session.
//Get api key here: https://www.cleverbot.com/api/.
func New(yourAPIKey string) *Session {
	values := &url.Values{}
	values.Set("key", yourAPIKey)
	values.Set("wrapper", "cleverbot-go")

	return &Session{
		&http.Client{},
		values,
		make(map[string]interface{}),
	}
}

// Ask asks cleverbot a question.
func (s *Session) Ask(question string) (string, error) {
	s.values.Set("input", question)

	// Make the actual request
	req, err := http.NewRequest("GET", apiURL+s.values.Encode(), nil)
	if err != nil {
		return "", err
	}

	// Headers
	req.Header.Set("User-Agent", "cleverbot-go https://github.com/ugjka/cleverbot-go")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Pragma", "no-cache")

	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Check for errors
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		return "", errors.New("Cleverbot API: key not valid")
	case http.StatusNotFound:
		return "", errors.New("Cleverbot API: not found")
	case http.StatusRequestEntityTooLarge:
		return "", errors.New("Cleverbot API: request too large. Please limit requests to 8KB")
	case http.StatusBadGateway:
		return "", errors.New("Cleverbot API: unable to get reply from API server, please contact us")
	case http.StatusGatewayTimeout:
		return "", errors.New("Cleverbot API: unable to get reply from API server, please contact us")
	case http.StatusServiceUnavailable:
		return "", errors.New("Cleverbot API: Too many requests from client")
	default:
		if resp.StatusCode != http.StatusOK {
			return "", errors.New("Cleverbot API: Error, " + string(resp.StatusCode) + " response code")
		}
	}
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if err := json.Unmarshal(body, &s.decoder); err != nil {
		return "", err
	}
	if _, ok := s.decoder["output"].(string); !ok {
		return "", errors.New("Cleverbot API: 'output' is not a string")
	}
	if _, ok := s.decoder["cs"].(string); !ok {
		return "", errors.New("Cleverbot API: 'cs' is not a string")
	}
	//Set session context id
	s.values.Set("cs", s.decoder["cs"].(string))

	return s.decoder["output"].(string), nil
}
