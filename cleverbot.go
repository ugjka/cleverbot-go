// Package cleverbot a wrapper for cleverbot.com
// To get a new session call New() and to ask call Session.Ask(question)
package cleverbot

import (
	"bytes"
	"crypto/md5"
	"errors"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

var warning = `WARNING: cleverbot-go is against cleverbot.com terms of service, use at your own discretion!`

func init() {
	fmt.Println(warning)
}

var (
	host     = "www.cleverbot.com"
	protocol = "http://"
	resource = "/webservicemin?uc="
	apiURL   = protocol + host + resource
)

//Holds history
type vtext []string

//Session is cleverbot session
type Session struct {
	V         vtext
	History   int
	ContextID string
	Client    *http.Client
	APIID     string
}

//New creates a new session
func New() (*Session, error) {
	apiID, err := getAPIID(protocol + host)
	if err != nil {
		return nil, err
	}
	return &Session{
		make([]string, 1),
		16,
		"",
		&http.Client{},
		apiID,
	}, nil
}

// Ask cleverbot a question
func (s *Session) Ask(q string) (string, error) {
	q = url.QueryEscape(q)
	var push string
	if s.ContextID == "" {
		push = "stimulus=" + q + "&cb_settings_language=en&cb_settings_scripting=no&islearning=1&icognoid=wsf"
	} else {
		push = "stimulus=" + q + "&" + s.V.history() + "cb_settings_language=en&cb_settings_scripting=no&sessionid=" + s.ContextID + "&islearning=1&icognoid=wsf"
	}

	// A hash of part of the payload, cleverbot needs this for some reason
	digestTxt := push[9:35]
	tokenMd5 := md5.New()
	io.WriteString(tokenMd5, digestTxt)
	tokenbuf := hexDigest(tokenMd5)
	token := tokenbuf.String()
	push = push + "&icognocheck=" + token

	// Make the actual request
	req, err := http.NewRequest("POST", apiURL+s.APIID+"&botapi=see%20www.cleverbot.com%2Fapis&", strings.NewReader(push))
	if err != nil {
		return "", err
	}

	// Headers and a cookie, which cleverbot again will not work without
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.90 Safari/537.36")
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Host", host)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Charset", "ISO-8859-1,utf-8;q=0.7,*;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,lv;q=0.6,ru;q=0.4,da;q=0.2")
	req.Header.Set("Referer", protocol+host+"/")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cookie", "XVIS=TEI939AFFIAGAYQZ")

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Get session ID
	s.ContextID = resp.Header.Get("CBCONVID")

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	//Parse response
	answer := ""
	for i, by := range body {
		if by == byte(13) {
			res := body[:i]
			answer = string(res)
			break
		}
	}

	//Detect Api changes
	if strings.Contains(answer, "<html>") {
		s.APIID, err = getAPIID(protocol + host)
		if err != nil {
			return "", err
		}
		return "Error: ask again!", nil
	}

	//Vtext
	s.V = append(s.V, q)
	s.V = append(s.V, url.QueryEscape(answer))

	//Stop history from growing endlessly
	if len(s.V) > s.History {
		s.V = s.V[2:]
	}

	return answer, nil
}

func hexDigest(hash hash.Hash) bytes.Buffer {
	var hexsum bytes.Buffer
	for _, i := range hash.Sum(nil) {
		fmt.Fprintf(&hexsum, "%02x", i)
	}
	return hexsum
}

func (v vtext) string() string {
	res := ""
	for i, v := range v {
		if i > 0 {
			res = res + " "
		}
		res = res + v
	}
	return res
}

//Generate vText strings
func (v vtext) history() string {
	if len(v) == 0 {
		return ""
	}
	result := ""
	for i, j := len(v), 2; i > 1; i, j = i-1, j+1 {
		result = result + "vText" + strconv.Itoa(j) + "=" + v[i-1:i].string() + "&"
	}
	return result
}

func getter(url string) (string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	return string(body), nil
}

func getAPIID(url string) (string, error) {
	body, err := getter(url)
	if err != nil {
		return "", err
	}
	getjs := regexp.MustCompile("(conversation-social-min.js\\?\\d+)")
	jsfile := getjs.FindStringSubmatch(body)
	if len(jsfile) < 2 {
		return "", errors.New("No regex matches for conversation-social-min.js in index.html ")
	}
	body, err = getter(url + "/extras/" + jsfile[1])
	if err != nil {
		return "", err
	}
	getapi := regexp.MustCompile("\"uc=(\\d+)&botapi=\"")
	apiID := getapi.FindStringSubmatch(body)
	if len(apiID) < 2 {
		return "", errors.New("No regex matches for api Id in conversation-social-min.js")
	}
	return apiID[1], nil
}
