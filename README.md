# WARNING: this library uses an unofficial api and is against cleverbot.com TOS, use the official cleverbot api instead https://www.cleverbot.com/api/

# cleverbot-go [![Build Status](https://travis-ci.org/ugjka/cleverbot-go.svg?branch=master)](https://travis-ci.org/ugjka/cleverbot-go) [![GoDoc](https://godoc.org/github.com/ugjka/cleverbot-go?status.svg)](https://godoc.org/github.com/ugjka/cleverbot-go) [![Go Report Card](https://goreportcard.com/badge/github.com/ugjka/cleverbot-go)](https://goreportcard.com/report/github.com/ugjka/cleverbot-go)
Cleverbot wrapper written in Go https://godoc.org/github.com/ugjka/cleverbot-go

## Example

```go
session, _ := cleverbot.New()
answer, _ := session.Ask("How are you?")
```
