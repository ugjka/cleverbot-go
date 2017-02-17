package main

import (
	"bufio"
	"fmt"
	"log"
	"os"

	cleverbot "github.com/ugjka/cleverbot-go"
)

func main() {
	//Get your api key here: http://www.cleverbot.com/api/
	cb := cleverbot.New("YOURAPIKEY")

	scanner := bufio.NewScanner(os.Stdin)
	fmt.Println("Press CTRL-C to exit.")

	// Start chat loop.
	for scanner.Scan() {
		response, err := ask(cb, scanner.Text())
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println("Cleverbot: " + response)
	}

	if err := scanner.Err(); err != nil {
		fmt.Fprintln(os.Stderr, "Error:", err)
	}
}

func ask(cb *cleverbot.Session, input string) (response string, err error) {
	return cb.Ask(input)
}
