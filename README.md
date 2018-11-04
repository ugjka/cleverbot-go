# cleverbot-go
[![Build Status](https://travis-ci.org/ugjka/cleverbot-go.svg?branch=master)](https://travis-ci.org/ugjka/cleverbot-go)
[![GoDoc](https://godoc.org/github.com/ugjka/cleverbot-go?status.svg)](https://godoc.org/github.com/ugjka/cleverbot-go)
[![Go Report Card](https://goreportcard.com/badge/github.com/ugjka/cleverbot-go)](https://goreportcard.com/report/github.com/ugjka/cleverbot-go)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)
[![Donate](https://dl.ugjka.net/Donate-PayPal-green.svg)](https://www.paypal.me/ugjka)

Cleverbot api wrapper written in Go

## Installation
    go get -u github.com/ugjka/cleverbot-go

Check out the [godoc](https://godoc.org/github.com/ugjka/cleverbot-go), for methods

## Example

GET API KEY HERE: http://www.cleverbot.com/api/
```go
package main

import (
	"fmt"

	"github.com/ugjka/cleverbot-go"
)

func main() {
	session := cleverbot.New("YOURAPIKEY")
	answer, _ := session.Ask("Hi, How are you?")
	fmt.Println(answer)
}
```
