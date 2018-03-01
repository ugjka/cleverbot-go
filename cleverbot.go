// Package cleverbot is a wrapper for cleverbot.com api.
// Get the official API Key here http://www.cleverbot.com/api/ .
package cleverbot

import (
	"encoding/json"
	"fmt"
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

type cleverErr struct {
	err    string
	origin string
}

func (e *cleverErr) Error() string {
	return e.err + " | " + e.origin
}

// Errors.
var (
	// ErrKeyNotValid is returned when API key is not valid.
	ErrKeyNotValid = &cleverErr{err: "cleverbot api: key not valid"}
	// ErrAPINotFound is returned when API returns 404.
	ErrAPINotFound = &cleverErr{err: "cleverbot api: not found"}
	// ErrRequestTooLarge is returned when GET request to the api exceeds 16K.
	ErrRequestTooLarge = &cleverErr{err: "cleverbot api: request too large. Please limit requests to 8KB"}
	// ErrNoReply is returned when api is down.
	ErrNoReply = &cleverErr{err: "cleverbot api: unable to get reply from API server, please contact us"}
	// ErrTooManyRequests is returned when there is too many requests made to the api.
	ErrTooManyRequests = &cleverErr{err: "cleverbot api: too many requests from client"}
	// ErrStatusNotOK is returned when statuscode is not 200
	ErrStatusNotOK = &cleverErr{err: "cleverbot api: response status not ok"}
	// ErrInvalidJSON is returned when the cleverbot server sends malformed json
	ErrInvalidJSON = &cleverErr{err: "cleverbot api: server sent malformed JSON"}
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
	mu     sync.Mutex
	Client *http.Client
	//Request parameters
	Values *url.Values
	//Decoded json
	Decoded map[string]interface{}
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
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Values.Set("input", question)
	// Clear previous json, just in case
	s.clear()
	// Prepare the request.
	req, err := http.NewRequest("GET", apiURL+s.Values.Encode(), nil)
	if err != nil {
		return "", err
	}

	// Headers.
	req.Header.Set("User-Agent", "cleverbot-go https://github.com/ugjka/cleverbot-go")

	// Make the request
	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	// Check for errors.
	switch resp.StatusCode {
	case http.StatusUnauthorized:
		ErrKeyNotValid.origin = fmt.Sprintf("status code %d", resp.StatusCode)
		return "", ErrKeyNotValid
	case http.StatusNotFound:
		ErrAPINotFound.origin = fmt.Sprintf("status code %d", resp.StatusCode)
		return "", ErrAPINotFound
	case http.StatusRequestEntityTooLarge:
		ErrRequestTooLarge.origin = fmt.Sprintf("status code %d", resp.StatusCode)
		return "", ErrRequestTooLarge
	case http.StatusBadGateway:
		ErrNoReply.origin = fmt.Sprintf("status code %d", resp.StatusCode)
		return "", ErrNoReply
	case http.StatusGatewayTimeout:
		ErrNoReply.origin = fmt.Sprintf("status code %d", resp.StatusCode)
		return "", ErrNoReply
	case http.StatusServiceUnavailable:
		ErrTooManyRequests.origin = fmt.Sprintf("status code %d", resp.StatusCode)
		return "", ErrTooManyRequests
	default:
		if resp.StatusCode != http.StatusOK {
			ErrStatusNotOK.origin = fmt.Sprintf("status code %d", resp.StatusCode)
			return "", ErrStatusNotOK
		}
	}

	if err := json.NewDecoder(resp.Body).Decode(&s.Decoded); err != nil {
		ErrInvalidJSON.origin = err.Error()
		return "", ErrInvalidJSON
	}
	if _, ok := s.Decoded["output"].(string); !ok {
		return "", fmt.Errorf("Cleverbot API: 'output' does not exist or is not a string")
	}
	if _, ok := s.Decoded["cs"].(string); !ok {
		return "", fmt.Errorf("Cleverbot API: 'cs' does not exist or is not a string")
	}
	// Set session context id.
	s.Values.Set("cs", s.Decoded["cs"].(string))

	return s.Decoded["output"].(string), nil
}

// Reset resets the session, meaning a new Ask() call will appear as new conversation from bots point of view.
func (s *Session) Reset() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.Values.Del("cs")
	// Clear the json map
	s.clear()
}

// InteractionCount gets the count of interactions that have happened between the bot and user.
// Returns -1 if interactions_count is missing or parsing failed.
func (s *Session) InteractionCount() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.Decoded["interaction_count"].(string); ok {
		if count, err := strconv.Atoi(v); err == nil {
			return count
		}
	}
	return -1
}

// TimeElapsed returns approximate duration since conversation started.
// Returns -1 seconds if time_elapsed is not found or parsing failed.
func (s *Session) TimeElapsed() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.Decoded["time_elapsed"].(string); ok {
		if dur, err := time.ParseDuration(v + "s"); err == nil {
			return dur
		}
	}
	return time.Second * -1
}

// TimeTaken returns the duration the bot took to respond.
// Returns -1 second if time_taken not found or parsing failed.
func (s *Session) TimeTaken() time.Duration {
	s.mu.Lock()
	defer s.mu.Unlock()
	if v, ok := s.Decoded["time_taken"].(string); ok {
		if dur, err := time.ParseDuration(v + "ms"); err == nil {
			return dur
		}
	}
	return time.Second * -1
}

// History returns an array of QApair of upto 100 interactions that have happened in Session.
func (s *Session) History() QAPairs {
	s.mu.Lock()
	defer s.mu.Unlock()
	var qa []QAPair
	for i := 1; ; i++ {
		if v, ok := s.Decoded[fmt.Sprintf("interaction_%d_other", i)].(string); ok && v != "" {
			qa = append([]QAPair{{s.Decoded[fmt.Sprintf("interaction_%d", i)].(string),
				s.Decoded[fmt.Sprintf("interaction_%d_other", i)].(string)}}, qa...)
		} else {
			return qa
		}
	}
}

func (s *Session) clear() {
	// Clear the map
	for k := range s.Decoded {
		delete(s.Decoded, k)
	}
}

// Wackiness varies Cleverbot’s reply from sensible to wacky.
// Accepted values between 0 and 100
func (s *Session) Wackiness(val uint8) {
	if val > 100 {
		val = 100
	}
	s.Values.Set("cb_settings_tweak1", fmt.Sprint(val))
}

// Talkativeness varies Cleverbot’s reply from shy to talkative.
// Accepted values between 0 and 100
func (s *Session) Talkativeness(val uint8) {
	if val > 100 {
		val = 100
	}
	s.Values.Set("cb_settings_tweak2", fmt.Sprint(val))
}

// Attentiveness varies Cleverbot’s reply from self-centred to attentive.
// Accepted values between 0 and 100
func (s *Session) Attentiveness(val uint8) {
	if val > 100 {
		val = 100
	}
	s.Values.Set("cb_settings_tweak3", fmt.Sprint(val))
}
