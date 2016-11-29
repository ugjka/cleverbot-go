// Very simple go cleverbot wrapper
// To get a new session call New() and to ask call Session.Ask(question)
package cleverbot

import (
	"bytes"
	"crypto/md5"
	"fmt"
	"hash"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

var (
	HOST     = "www.cleverbot.com"
	PROTOCOL = "http://"
	RESOURCE = "/webservicemin?uc=321&"
	API_URL  = PROTOCOL + HOST + RESOURCE
)

//Holds history
type vtext []string

type Session struct {
	V          vtext
	History    int
	Context_id string
	Client     *http.Client
}

// Creates a new session
func New() *Session {
	return &Session{
		make([]string, 1),
		16,
		"",
		&http.Client{},
	}
}

// Ask cleverbot a question
func (s *Session) Ask(q string) (string, error) {
	q = url.QueryEscape(q)
	push := ""
	if s.Context_id == "" {
		push = "stimulus=" + q + "&cb_settings_language=en&cb_settings_scripting=no&islearning=1&icognoid=wsf"
	} else {
		push = "stimulus=" + q + "&" + s.V.history() + "cb_settings_language=en&cb_settings_scripting=no&sessionid=" + s.Context_id + "&islearning=1&icognoid=wsf"
	}

	// A hash of part of the payload, cleverbot needs this for some reason
	digest_txt := push[9:35]
	tokenMd5 := md5.New()
	io.WriteString(tokenMd5, digest_txt)
	tokenbuf := hexDigest(tokenMd5)
	token := tokenbuf.String()
	push = push + "&icognocheck=" + token

	// Make the actual request
	req, err := http.NewRequest("POST", API_URL, strings.NewReader(push))
	if err != nil {
		return "", err
	}

	// Headers and a cookie, which cleverbot again will not work without
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/54.0.2840.90 Safari/537.36")
	req.Header.Set("Content-Type", "text/plain;charset=UTF-8")
	req.Header.Set("Host", HOST)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Charset", "ISO-8859-1,utf-8;q=0.7,*;q=0.7")
	req.Header.Set("Accept-Language", "en-US,en;q=0.8,lv;q=0.6,ru;q=0.4,da;q=0.2")
	req.Header.Set("Referer", PROTOCOL+HOST+"/")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cookie", "XVIS=TEI939AFFIAGAYQZ")

	resp, err := s.Client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	// Get session ID
	s.Context_id = resp.Header.Get("CBCONVID")

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
