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

//Api adress
var (
	host     = "www.cleverbot.com"
	protocol = "http://"
	resource = "/getreply?"
	apiURL   = protocol + host + resource
)

//Errors
var (
	//ErrKeyNotValid is returned when API key is not valid
	ErrKeyNotValid = errors.New("Cleverbot API: key not valid")
	//ErrAPINotFound is returned when API returns 404
	ErrAPINotFound = errors.New("Cleverbot API: not found")
	//ErrRequestTooLarge is returned when GET request to the api exceeds 16K
	ErrRequestTooLarge = errors.New("Cleverbot API: request too large. Please limit requests to 8KB")
	//ErrNoReply is returned when api is down
	ErrNoReply = errors.New("Cleverbot API: unable to get reply from API server, please contact us")
	//ErrTooManyRequests is returned when there is too many requests made to the api
	ErrTooManyRequests = errors.New("Cleverbot API: Too many requests from client")
	//ErrUnknown is returned when response status code is not 200
	ErrUnknown = errors.New("Cleverbot API: Response is not 200, unknown error")
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
		return "", ErrKeyNotValid
	case http.StatusNotFound:
		return "", ErrAPINotFound
	case http.StatusRequestEntityTooLarge:
		return "", ErrRequestTooLarge
	case http.StatusBadGateway:
		return "", ErrNoReply
	case http.StatusGatewayTimeout:
		return "", ErrNoReply
	case http.StatusServiceUnavailable:
		return "", ErrTooManyRequests
	default:
		if resp.StatusCode != http.StatusOK {
			return "", ErrUnknown
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
