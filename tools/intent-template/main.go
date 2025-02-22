package main

import (
	"fmt"
	"log"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type IntentTemplate struct {
	Intent     string                    `json:"intent"`
	Attributes []*AttributeVariablePair  `json:"attributes,omitempty"`
	Slots      []*IntentSlotVariablePair `json:"slots,omitempty"`
}

type AttributeVariablePair struct {
	Attribute string `json:"attribute"`
	Variable  string `json:"variable"`
}

type IntentSlotVariablePair struct {
	Name     string `json:"name"`
	Variable string `json:"variable"`
	Type     string `json:"type"`
	Default  string `json:"default,omitempty"`
}

func (i *IntentSlotVariablePair) IsValidType() bool {
	switch i.Type {
	case "string":
		return true
	case "int":
		return true
	case "float64":
		return true
	default:
		return false
	}
}

func (i *IntentTemplate) GenerateCode() (string, error) {
	var out string

	diminutiveIntentName := strings.ToLower(string(i.Intent[0]))
	diminutiveIntentName += i.Intent[1:]
	structName := diminutiveIntentName + "Handler"

	out = fmt.Sprintf(`
// -----------------------
type %s struct{}

func (h *%s) CanHandle(input askgo.HandlerInput) bool {
    request := input.GetRequest()
    return request.Intent.Name == "%s"
}

func (h *%s) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
    request := input.GetRequest()
    builder := input.GetResponse().WithShouldEndSession(false)
    attributes := input.GetContext().Value(&attributeContext).(*Attributes)

    log.Printf("%s requestId=%%s, sessionId=%%s", request.RequestID, attributes.sessionID)

    userID := input.GetRequestEnvelope().Context.System.User.UserID
    deviceID := input.GetRequestEnvelope().Context.System.Device.DeviceID

`, structName, structName, i.Intent, structName, structName)

	if len(i.Attributes) > 0 {
		out += `    // Each of these should be in the Attributes struct definition\n`
		for _, a := range i.Attributes {
			out += fmt.Sprintf(`    %s := attributes.%s
`, a.Variable, a.Attribute)
		}
		out += "\n"
	}

	if len(i.Slots) > 0 {
		for _, s := range i.Slots {
			if !s.IsValidType() {
				return "", fmt.Errorf("Variable %s is an invalid type %s", s.Variable, s.Type)
			}
			out += fmt.Sprintf("    var %s %s\n", s.Variable, s.Type)
			out += fmt.Sprintf("    var %sFound bool\n", s.Variable)
		}

		out += `
    if len(request.Intent.Slots) > 0 {
        for _, i := range request.Intent.Slots {
            if i.IsSlotValidValue() {
                switch i.Name {
`
		for _, s := range i.Slots {
			prefixSpaces := "                "
			switch s.Type {
			case "string":
				out += fmt.Sprintf("%scase \"%s\":\n", prefixSpaces, s.Name)
				out += fmt.Sprintf("%s    %s = i.Value\n", prefixSpaces, s.Variable)
				out += fmt.Sprintf("%s    %sFound = true\n", prefixSpaces, s.Variable)
			case "int":
				out += fmt.Sprintf("%scase \"%s\":\n", prefixSpaces, s.Name)
				out += fmt.Sprintf("%s    %s, err = strconv.Atoi(i.Value)\n", prefixSpaces, s.Variable)
				out += fmt.Sprintf("%s    if err != nil {\n", prefixSpaces)
				if s.Default != "" {
					out += fmt.Sprintf("%s        log.Printf(\"%s integer conversion error, defaulting to %s: %%v\", err)\n", prefixSpaces, s.Name, s.Default)
					out += fmt.Sprintf("%s        %s = %s\n", prefixSpaces, s.Variable, s.Default)
				} else {
					out += fmt.Sprintf("%s        log.Printf(\"%s integer conversion error: %%v\", err)\n", prefixSpaces, s.Name)
					out += fmt.Sprintf("%s        return builder.Speak(fmt.Sprintf(\"There was an error converting the value of %s - %%v\", err)), nil\n", prefixSpaces, s.Name)
				}
				out += fmt.Sprintf("%s    }\n", prefixSpaces)
				out += fmt.Sprintf("%s    %sFound = true\n", prefixSpaces, s.Variable)
			case "float64":
				out += fmt.Sprintf("%scase \"%s\":\n", prefixSpaces, s.Name)
				out += fmt.Sprintf("%s    %s, err = strconv.ParseFloat(i.Value, 64)\n", prefixSpaces, s.Variable)
				out += fmt.Sprintf("%s    if err != nil {\n", prefixSpaces)
				if s.Default != "" {
					out += fmt.Sprintf("%s        log.Printf(\"%s floating point conversion error, defaulting to %s: %%v\", err)\n", prefixSpaces, s.Name, s.Default)
					out += fmt.Sprintf("%s        %s = %s\n", prefixSpaces, s.Variable, s.Default)
				} else {
					out += fmt.Sprintf("%s        log.Printf(\"%s floating point conversion error: %%v\", err)\n", prefixSpaces, s.Name)
					out += fmt.Sprintf("%s        return builder.Speak(fmt.Sprintf(\"There was an error converting the value of %s - %%v\", err)), nil\n", prefixSpaces, s.Name)
				}
				out += fmt.Sprintf("%s    }\n", prefixSpaces)
				out += fmt.Sprintf("%s    %sFound = true\n", prefixSpaces, s.Variable)
			}
		}
		out += `                }
                log.Printf("Valid value for slot %s: %s", i.Name, i.Value)
            } else {
                log.Printf("Value for slot %s is reportedly invalid [%s]", i.Name, i.Value)
            }
        }
    }
`
		out += `
    // Optional code to respond to missing slots
`
		for _, s := range i.Slots {
			out += fmt.Sprintf(`    if (!%sFound) {
        // Do something for %s value missing
    }
`, s.Variable, s.Name)
		}
	}

	out += `
    // Your logic code goes here

    var responseStr string

    // Respond to the user
    return builder.Speak(responseStr), nil
}
`
	return out, nil
}

var helpString = `
Syntax: intent-template <YAML file>
YAML file should conform to the following schema:

intent: BudgetAnalyzeTransactionsIntent
attributes:
- attribute: BudgetID
  variable: budgetID
- attribute: LastUserID
  variable: lastUID
slots:
- name: ItemType
  variable: itemType
  type: string    # (string, int, float64 supported)
- name: TransactionCount
  variable: txnCount
  type: int
- name: Amount
  variable: amount
  type: float64
  default: "0.00"
`

func do_help() {
	fmt.Println(helpString)
}

func main() {
	if len(os.Args) < 2 {
		do_help()
		os.Exit(1)
	}

	f, err := os.Open(os.Args[1])
	if err != nil {
		log.Fatalf("Error opening YAML file %s: %v", os.Args[1], err)
	}
	defer f.Close()

	obj := new(IntentTemplate)
	dec := yaml.NewDecoder(f)
	err = dec.Decode(obj)
	if err != nil {
		log.Fatalf("Error decoding YAML document: %v", err)
	}

	// Process contents and generate Intent code
	out, err := obj.GenerateCode()
	if err != nil {
		log.Fatalf("GenerateCode() threw an error: %v", err)
	}

	fmt.Println(out)
}
