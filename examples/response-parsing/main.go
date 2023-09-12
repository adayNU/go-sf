package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/adayNU/go-sf/pkg/constants"
	"github.com/fabiustech/anthropic"
)

// Recipe from https://cooking.nytimes.com/recipes/1017354-pimento-cheese.
const recipe = `A decidedly Southern spread with Northern roots, pimento cheese is a simple mix of Cheddar, red bell pepper and mayonnaise that can be found on sandwiches or served as a spread for crackers from work sites to garden parties across the 16 states below the Mason-Dixon line. This recipe came to The Times from the Charleston, S.C.-bred cookbook authors Matt Lee and Ted Lee.

INGREDIENTS
Yield:
About 1.5 cups, enough for 4 sandwiches
8 ounces extra-sharp Cheddar, grated with a food processor or hand grater (not pre-grated)
1/4 cup softened cream cheese (2 ounces), pulled into several pieces
Scant 1/2 cup jarred pimento or other roasted red peppers (from a 7-ounce jar), finely diced
3 tablespoons Duke’s, Hellmann’s or other high-quality store-bought mayonnaise
1/2 teaspoon red-pepper flakes
Salt and black pepper to taste

PREPARATION
Step 1
In a large mixing bowl, place the Cheddar in an even layer. Scatter the cream cheese, pimentos, mayonnaise, red-pepper flakes and salt and pepper over the Cheddar. Using a spatula, mix the pimento cheese until it is smooth and spreadable.

Step 2
Transfer the pimento cheese to a bowl or container, cover tightly, and store in the refrigerator for up to 1 week.`

const xmlPrefix = `<shopping-list>`

// prompt returns a list of messages to send to the API.
func prompt(recipe string) []*anthropic.Message {
	return []*anthropic.Message{
		{
			Text: `I have been given a recipe for which I would like to generate a shopping list of the items I'd need to buy in order to prepare it.

I would like you to output the item and the quantity of that item that I would need to buy in the following XML format:

<shopping-list>
	<item>
		<name>item name</name>
		<quantity>item quantity</quantity>
	</item>
</shopping-list>

Do you understand the task and output structure? I will provide the recipe once you confirm.`,
			UserType: anthropic.UserTypeHuman,
		},
		{
			Text:     "Yes, I understand the task and output structure (valid XML). Please provide the recipe.",
			UserType: anthropic.UserTypeAssistant,
		},
		{
			Text: fmt.Sprintf(`The <recpie> is below:

<recipe>
%s
</recipe>`, recipe),
			UserType: anthropic.UserTypeHuman,
		},
		{
			Text:     xmlPrefix,
			UserType: anthropic.UserTypeAssistant,
		},
	}
}

// Item represents an item in a shopping list.
type Item struct {
	Name     string `xml:"name"`
	Quantity string `xml:"quantity"`
}

// ShoppingList represents a shopping list.
type ShoppingList struct {
	xml.Name `xml:"shopping-list"`
	Items    []*Item `xml:"item"`
}

func main() {
	var key, ok = os.LookupEnv(constants.AnthropicAPIKey)
	if !ok {
		panic(fmt.Sprintf("environment variable %s not set", constants.AnthropicAPIKey))
	}

	var c = anthropic.NewClient(key)
	var resp, err = c.NewCompletion(context.Background(), &anthropic.Request{
		Prompt:            anthropic.NewPromptFromMessages(prompt(recipe)),
		Model:             anthropic.Claude,
		MaxTokensToSample: 500,
		Temperature:       anthropic.Optional(0.0),
	})
	if err != nil {
		panic(err)
	}

	var list = &ShoppingList{}
	// We need to append the XML prefix to the response because we've "put words in the model's mouth" by
	// including the prefix in the prompt.
	// See: https://docs.anthropic.com/claude/docs/human-and-assistant-formatting#use-human-and-assistant-to-put-words-in-claudes-mouth
	if err = xml.Unmarshal([]byte(xmlPrefix+resp.Completion), list); err != nil {
		panic(err)
	}

	for _, item := range list.Items {
		fmt.Printf("Item: %s \n\t%s\n", item.Name, item.Quantity)
	}
}
