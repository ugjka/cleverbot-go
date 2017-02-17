# cleverbot-go [![Build Status](https://travis-ci.org/ugjka/cleverbot-go.svg?branch=master)](https://travis-ci.org/ugjka/cleverbot-go) [![GoDoc](https://godoc.org/github.com/ugjka/cleverbot-go?status.svg)](https://godoc.org/github.com/ugjka/cleverbot-go) [![Go Report Card](https://goreportcard.com/badge/github.com/ugjka/cleverbot-go)](https://goreportcard.com/report/github.com/ugjka/cleverbot-go)
Cleverbot wrapper written in Go

## Example

GET API KEY HERE: http://www.cleverbot.com/api/
```go
session := cleverbot.New("YOURAPIKEY")
answer, _ := session.Ask("How are you?")
```
