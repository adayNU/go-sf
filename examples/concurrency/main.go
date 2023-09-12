package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"sync"

	"github.com/adayNU/go-sf/internal/constants"
	"github.com/fabiustech/anthropic"
)

// SemaphoreClient is a semaphore for anthropic clients.
type SemaphoreClient chan *anthropic.Client

// NewSemaphoreClient creates a new semaphore for anthropic clients.
func NewSemaphoreClient(key string, size int) SemaphoreClient {
	var s = make(SemaphoreClient, size)
	for i := 0; i < size; i++ {
		s <- anthropic.NewClient(key)
	}

	return s
}

var workToBeDone = []anthropic.Prompt{
	anthropic.NewPromptFromString("Generate a shopping list to cook a thanksgiving dinner."),
	anthropic.NewPromptFromString("Output 30 different golang packages in a bulleted list."),
	anthropic.NewPromptFromString("Write a poem about why writing Leveraging LLMs in Go Applications is great."),
	anthropic.NewPromptFromString("What is the state capital of California?"),
	anthropic.NewPromptFromString("What is the state capital of New York?"),
	anthropic.NewPromptFromString(`Write a film script with three characters: Darius, a young journalist who is trying to appear smart and cool but is trying too hard; Sandra, a newly-famous director in her thirties who doesn't want to be here and speaks laconically; and George, a retired detective who doesn’t know anything about Darius and Sandra, but is observing them, and ends up inserting himself into their conversation. Darius just met Sandra for the interview. They are all at a small diner having breakfast; Darius and Sandra are at the same table, with George nearby.

Make it about 500 words and take it slow. No need for scene direction, just use prompts like “Darius:“, “Sandra:“, “George:” for the script.`),
}

func main() {
	var key, ok = os.LookupEnv(constants.AnthropicAPIKey)
	if !ok {
		panic(fmt.Sprintf("environment variable %s not set", constants.AnthropicAPIKey))
	}

	var wg = sync.WaitGroup{}
	var s = NewSemaphoreClient(key, 3)
	defer close(s)

	var results = make([]string, len(workToBeDone))

	for i, prompt := range workToBeDone {
		wg.Add(1)
		go func(p anthropic.Prompt, index int) {
			defer func() {
				log.Printf("finished task %v", index)
				wg.Done()
			}()

			var c = <-s
			defer func() { s <- c }()

			log.Printf("starting task %v", index)

			var resp, err = c.NewCompletion(context.Background(), &anthropic.Request{
				Prompt:            p,
				Model:             anthropic.Claude,
				MaxTokensToSample: 1000,
				Temperature:       anthropic.Optional(0.0),
			})
			if err != nil {
				panic(err)
			}

			results[index] = resp.Completion
		}(prompt, i)
	}

	wg.Wait()

	for i, result := range results {
		fmt.Printf("Result %v: %s\n\n", i+1, result)
	}
}
