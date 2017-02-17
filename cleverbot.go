// Package cleverbot a wrapper for cleverbot.com
// To get a new session call New("YOURAPIKEY") and to ask call Session.Ask(question)
// Get the official API Key here http://www.cleverbot.com/api/
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

//Session is cleverbot session
type Session struct {
	client  *http.Client
	values  *url.Values
	decoder map[string]interface{}
}

//New creates a new session
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

// Ask cleverbot a question
func (s *Session) Ask(question string) (string, error) {
	s.values.Set("input", question)

	// Make the actual request
	req, err := http.NewRequest("GET", apiURL+s.values.Encode(), nil)
	if err != nil {
		return "", err
	}

	// Headers
	req.Header.Set("User-Agent", "cleverbot-go https://github.com/ugjka/cleverbot-go")
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Host", host)
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
		return "", errors.New("401: unauthorised due to missing or invalid API key or POST request, the Cleverbot API only accepts GET requests")
	case http.StatusNotFound:
		return "", errors.New("404: API not found")
	case http.StatusRequestEntityTooLarge:
		return "", errors.New("413: request too large if you send a request over 16Kb")
	case http.StatusBadGateway:
		return "", errors.New("502: unable to get reply from API server, please contact us")
	case http.StatusGatewayTimeout:
		return "", errors.New("504: unable to get reply from API server, please contact us")
	case http.StatusServiceUnavailable:
		return "", errors.New("503: too many requests from a single IP address or API key")
	default:
		if resp.StatusCode != http.StatusOK {
			return "", errors.New("Got " + string(resp.StatusCode) + " response code")
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
		return "", errors.New("output: not a string")
	}
	if _, ok := s.decoder["cs"].(string); !ok {
		return "", errors.New("cs: not a string")
	}
	//Set session context id
	s.values.Set("cs", s.decoder["cs"].(string))

	return s.decoder["output"].(string), nil
}
