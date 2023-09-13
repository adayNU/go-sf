package main

import (
	"context"
	"encoding/xml"
	"fmt"
	"os"

	"github.com/adayNU/go-sf/internal/constants"
	"github.com/adayNU/go-sf/internal/parse"
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
// I've explicitly made the described schema incorrect to try and cause a parsing error.
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
		// This message is intended to represent a malformed response from Claude.
		{
			Text: `<shopping-list>
	<shopping-list>
  <item>
    <name>extra-sharp Cheddar</name>
    <quantity>8 ounces</quantity>
  </item>
  <item>
    <name>cream cheese</name> 
    <quantity>1/4 cup (2 ounces)</quantity>
  </item>
  <item>
    <name>jarred pimento or other roasted red peppers</name>
    <quantity>Scant 1/2 cup</quantity> 
  </item>
  <items>
    <name>Duke's, Hellmann's or other mayonnaise</name>
    <quantity>3 tablespoons</quantity>
  </item>
  <item>
    <name>red pepper flakes</name>
    <quantity>1/2 teaspoon</quantity>
  </item>
  <item>
    <name>salt</name>
    <quantity>to taste</quantity>
  </item>
  <item>
    <name>black pepper</name
	<quantity>to taste</quantity>
  </item>
</shopping-list>`,
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

	var msgs = prompt(recipe)

	var c = anthropic.NewClient(key)
	c.Debug() // This will log the prompt to stdout.

	var p = parse.NewRetryParser(c, ShoppingList{})
	var list, err = p.ParseXMLResponse(context.Background(), msgs, xmlPrefix)
	if err != nil {
		panic(err)
	}

	for _, item := range list.Items {
		fmt.Printf("Item: %s \n\t%s\n", item.Name, item.Quantity)
	}
}
