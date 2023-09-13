package parse

import (
	"context"
	"encoding/xml"
	"fmt"
	"log"

	"github.com/fabiustech/anthropic"
)

const (
	defaultMaxTokens  = 5000
	defaultMaxRetries = 2
)

// NewRetryParser returns a new RetryParser with default values.
func NewRetryParser[T any](claude *anthropic.Client, val T) *RetryParser[T] {
	return &RetryParser[T]{
		claude:     claude,
		maxRetries: defaultMaxRetries,
		maxTokens:  defaultMaxTokens,
		val:        val,
	}
}

// RetryParser represents a parser that asks Claude to regenerate based on the parsing error message
// and then continues to try to parse its subsequent responses at most |maxRetries| times.
type RetryParser[T any] struct {
	claude                *anthropic.Client
	maxRetries, maxTokens int
	val                   T
}

// ParseXMLResponse parses the response from Claude and returns the parsed value, retrying on error.
// It assumes that the response is in XML format, and Claudes response is the last message in |msg|s.
// If you "put words in the model's mouth" by including the XML prefix in the prompt, you should
// simply append the response to the last message before passing it to this function.
// Eg.
//
//	msgs[len(msgs)-1].Text += resp.Completion
func (r *RetryParser[T]) ParseXMLResponse(ctx context.Context, msgs []*anthropic.Message, prefix string) (*T, error) {
	if len(msgs) < 2 {
		return nil, fmt.Errorf("expected at least 2 messages, got %d", len(msgs))
	}

	var empty = r.val
	var unmarshalErr error

	for r.maxRetries >= 0 {
		var last = msgs[len(msgs)-1]

		if unmarshalErr = xml.Unmarshal([]byte(last.Text), &r.val); unmarshalErr == nil {
			return &r.val, nil
		}

		r.maxRetries--

		log.Printf("error unmarshaling response:, %s, tries left: %v", unmarshalErr.Error(), r.maxRetries)

		msgs = append(msgs, nonParsableClaudeResponse(unmarshalErr), &anthropic.Message{UserType: anthropic.UserTypeAssistant, Text: prefix})

		var resp, err = r.claude.NewCompletion(ctx, &anthropic.Request{
			Prompt:            anthropic.NewPromptFromMessages(msgs),
			Model:             anthropic.Claude,
			MaxTokensToSample: r.maxTokens,
			Temperature:       anthropic.Optional(0.0),
		})
		if err != nil {
			return nil, err
		}

		msgs[len(msgs)-1].Text += resp.Completion
		r.val = empty
	}

	return nil, fmt.Errorf("failed to parse response: %s", unmarshalErr.Error())
}

// nonParsableClaudeResponse returns a message to pass back to Claude when the response is not parsable.
func nonParsableClaudeResponse(err error) *anthropic.Message {
	return &anthropic.Message{
		UserType: anthropic.UserTypeHuman,
		Text: fmt.Sprintf(`It looks like your response was not in the correct format. I got the following <error> trying to unmarshal your response:

<error>
%s
</error>

Can you fix the issue with your previous response?`, err.Error()),
	}
}

// SetMaxRetries sets the maximum number of retries to parse the response.
func (r *RetryParser[T]) SetMaxRetries(max int) {
	r.maxRetries = max
}

// SetMaxTokens sets the maximum number of tokens to sample from Claude.
func (r *RetryParser[T]) SetMaxTokens(max int) {
	r.maxTokens = max
}
