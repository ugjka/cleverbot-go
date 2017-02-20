// Package cleverbot is a wrapper for cleverbot.com api.
// Get the official API Key here http://www.cleverbot.com/api/ .
package cleverbot

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"sync"
	"time"
)

// Api address.
var (
	host     = "www.cleverbot.com"
	protocol = "http://"
	resource = "/getreply?"
	apiURL   = protocol + host + resource
)

// Errors.
var (
	// ErrKeyNotValid is returned when API key is not valid.
	ErrKeyNotValid = errors.New("Cleverbot API: key not valid")
	// ErrAPINotFound is returned when API returns 404.
	ErrAPINotFound = errors.New("Cleverbot API: not found")
	// ErrRequestTooLarge is returned when GET request to the api exceeds 16K.
	ErrRequestTooLarge = errors.New("Cleverbot API: request too large. Please limit requests to 8KB")
	// ErrNoReply is returned when api is down.
	ErrNoReply = errors.New("Cleverbot API: unable to get reply from API server, please contact us")
	// ErrTooManyRequests is returned when there is too many requests made to the api.
	ErrTooManyRequests = errors.New("Cleverbot API: Too many requests from client")
	// ErrUnknown is returned when response status code is not 200.
	ErrUnknown = errors.New("Cleverbot API: Response is not 200, unknown error")
)

// QAPair contains question and answer pair strings of an interaction.
type QAPair struct {
	Question string
	Answer   string
}

// QAPairs is a slice of QAPair
type QAPairs []QAPair

func (q QAPair) String() string {
	return fmt.Sprintf("Question: %s Answer: %s", q.Question, q.Answer)
}

func (q QAPairs) String() string {
	result := ""
	for i, pair := range q {
		if i+1 == len(q) {
			result += fmt.Sprintf("%d: %s", i+1, pair)
		} else {
			result += fmt.Sprintf("%d: %s\n", i+1, pair)
		}
	}
	return result
}

// Session is a cleverbot session.
type Session struct {
	sync.Mutex
	client  *http.Client
	values  *url.Values
	decoder map[string]interface{}
}

// New creates a new session.
// Get api key here: https://www.cleverbot.com/api/ .
func New(yourAPIKey string) *Session {
	values := &url.Values{}
	values.Set("key", yourAPIKey)
	values.Set("wrapper", "cleverbot-go")

	return &Session{
		sync.Mutex{},
		&http.Client{},
		values,
		make(map[string]interface{}),
	}
}

// Ask asks cleverbot a question.
func (s *Session) Ask(question string) (string, error) {
	s.Lock()
	defer s.Unlock()
	s.values.Set("input", question)
	// Clear the map, just in case of some leftovers
	for k := range s.decoder {
		delete(s.decoder, k)
	}
	// Prepare the request.
	req, err := http.NewRequest("GET", apiURL+s.values.Encode(), nil)
	if err != nil {
		return "", err
	}

	// Headers.
	req.Header.Set("User-Agent", "cleverbot-go https://github.com/ugjka/cleverbot-go")

	// Make the request
	resp, err := s.client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Check for errors.
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
	// Set session context id.
	s.values.Set("cs", s.decoder["cs"].(string))

	return s.decoder["output"].(string), nil
}

// Reset resets the session, meaning a new Ask() call will appear as new conversation from bots point of view.
func (s *Session) Reset() {
	s.Lock()
	defer s.Unlock()
	s.values.Del("cs")
	// Clear the json map
	for k := range s.decoder {
		delete(s.decoder, k)
	}
}

// InteractionCount gets the count of interactions that have happened between the bot and user.
// Returns -1 if interactions_count is missing or an error occurred.
func (s *Session) InteractionCount() int {
	s.Lock()
	defer s.Unlock()
	if v, ok := s.decoder["interaction_count"].(string); ok {
		if count, err := strconv.Atoi(v); err == nil {
			return count
		}
	}
	return -1
}

// TimeElapsed returns approximate duration since conversation started. Returns -1 seconds on error or if time_elapsed is not found
func (s *Session) TimeElapsed() time.Duration {
	s.Lock()
	defer s.Unlock()
	if v, ok := s.decoder["time_elapsed"].(string); ok {
		if dur, err := time.ParseDuration(v + "s"); err == nil {
			return dur
		}
	}
	return time.Second * -1
}

// TimeTaken returns the duration the bot took to respond. Returns -1 second if not found or on error.
func (s *Session) TimeTaken() time.Duration {
	s.Lock()
	defer s.Unlock()
	if v, ok := s.decoder["time_taken"].(string); ok {
		if dur, err := time.ParseDuration(v + "ms"); err == nil {
			return dur
		}
	}
	return time.Second * -1
}

// History returns an array of QApair of upto 100 interactions that have happened in Session.
func (s *Session) History() QAPairs {
	s.Lock()
	defer s.Unlock()
	var qa []QAPair
	for i := 1; ; i++ {
		if v, ok := s.decoder["interaction_"+strconv.Itoa(i)+"_other"].(string); ok && v != "" {
			qa = append([]QAPair{{s.decoder["interaction_"+strconv.Itoa(i)].(string),
				s.decoder["interaction_"+strconv.Itoa(i)+"_other"].(string)}}, qa...)
		} else {
			return qa
		}
	}
}
