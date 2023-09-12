package main

import (
	"bytes"
	"context"
	"embed"
	"flag"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/adayNU/go-sf/pkg/constants"
	"github.com/fabiustech/anthropic"
)

//go:embed *.tmpl
var fs embed.FS

const (
	promptTemplate = "prompt.tmpl"
)

var (
	tmpl                  *template.Template
	first, last, industry *string
	concise               *bool
)

func init() {
	var err error
	tmpl, err = template.ParseFS(fs, promptTemplate)
	if err != nil {
		panic(err)
	}

	first = flag.String("first", "", "First name")
	last = flag.String("last", "", "Last name")
	industry = flag.String("industry", "", "Industry")
	concise = flag.Bool("concise", false, "Concise email")
	flag.Parse()
}

// Person represents a person you want to generate an email for.
type Person struct {
	FirstName, LastName, Industry string
	Concise                       bool
}

func main() {
	var key, ok = os.LookupEnv(constants.AnthropicAPIKey)
	if !ok {
		panic(fmt.Sprintf("environment variable %s not set", constants.AnthropicAPIKey))
	}

	if *first == "" || *last == "" || *industry == "" {
		panic("first, last, and industry are required")
	}

	var contact = &Person{
		FirstName: *first,
		LastName:  *last,
		Industry:  *industry,
		Concise:   *concise,
	}

	var buf = bytes.NewBuffer([]byte{})
	var err = tmpl.Execute(buf, contact)
	if err != nil {
		panic(err)
	}

	var c = anthropic.NewClient(key)
	c.Debug() // This will log the prompt to stdout.

	var resp *anthropic.Response
	resp, err = c.NewCompletion(context.Background(), &anthropic.Request{
		Prompt:            anthropic.NewPromptFromString(strings.TrimSpace(buf.String())),
		Model:             anthropic.Claude,
		MaxTokensToSample: 500,
		Temperature:       anthropic.Optional(0.0),
	})
	if err != nil {
		panic(err)
	}

	fmt.Printf("\n\n%s\n", resp.Completion)
}
