package main

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/fatih/structs"
	"github.com/spirilis/askgo"
	"github.com/spirilis/askgo/alexa"
)

func supportsDisplay(ctx alexa.Context) bool {
	_, found := ctx.System.Device.SupportedInterfaces["Display"]
	return found
}

//  -----------------------
var attributeContext struct{}

type unpackAttributes struct{}

func (h *unpackAttributes) Process(input askgo.HandlerInput) error {
	attributes := getAttributes(input)

	log.Printf("Got Attributes")

	input.SetContext(context.WithValue(input.GetContext(), &attributeContext, attributes))

	return nil
}

//  -----------------------
type saveAttributes struct{}

func (h *saveAttributes) Process(input askgo.HandlerInput, envelope *askgo.ResponseEnvelope) error {
	if !envelope.Response.ShouldSessionEnd {
		attributes, ok := input.GetContext().Value(&attributeContext).(*Attributes)

		if !ok {
			log.Printf("Error: Attributes not correct type")
		} else {
			envelope.SessionAttributes = structs.Map(attributes)
		}
	}

	return nil
}

//  -----------------------
type errorHandler struct{}

func (h *errorHandler) CanHandle(input askgo.HandlerInput) bool {
	return true
}
func (h *errorHandler) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
	request := input.GetRequest()
	builder := input.GetResponse().WithShouldEndSession(false)
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	log.Printf("ErrorHandler requestId=%s, sessionId=%s", request.RequestID, attributes.sessionID)

	return builder.Speak(helpMessage).Reprompt(helpMessage), nil
}

//  -----------------------
type helpHandler struct{}

func (h *helpHandler) CanHandle(input askgo.HandlerInput) bool {
	request := input.GetRequest()
	return request.Intent.Name == alexa.HelpIntent
}
func (h *helpHandler) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
	request := input.GetRequest()
	builder := input.GetResponse().WithShouldEndSession(false)
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	log.Printf("HelpHandler requestId=%s, sessionId=%s", request.RequestID, attributes.sessionID)

	return builder.Speak(helpMessage).Reprompt(helpMessage), nil
}

//  -----------------------
type exitHandler struct{}

func (h *exitHandler) CanHandle(input askgo.HandlerInput) bool {
	request := input.GetRequest()
	return request.Intent.Name == alexa.StopIntent ||
		request.Intent.Name == alexa.PauseIntent ||
		request.Intent.Name == alexa.CancelIntent
}
func (h *exitHandler) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
	request := input.GetRequest()
	builder := input.GetResponse().WithShouldEndSession(true)
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	log.Printf("ExitHandler requestId=%s, sessionId=%s", request.RequestID, attributes.sessionID)

	return builder.Speak(exitSkillMessage), nil
}

//  -----------------------
type sessionEndHandler struct{}

func (h *sessionEndHandler) CanHandle(input askgo.HandlerInput) bool {
	return input.GetRequest().Type == "SessionEndedRequest"
}
func (h *sessionEndHandler) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
	request := input.GetRequest()
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	log.Printf("SessionEnd requestId=%s, sessionId=%s", request.RequestID, attributes.sessionID)

	return input.GetResponse().WithShouldEndSession(true), nil
}

//  -----------------------
type launchHandler struct{}

func (h *launchHandler) CanHandle(input askgo.HandlerInput) bool {
	return input.GetRequest().Type == "LaunchRequest"
}
func (h *launchHandler) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
	request := input.GetRequest()
	response := input.GetResponse().WithShouldEndSession(false)
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	log.Printf("LaunchRequest requestId=%s, sessionId=%s", request.RequestID, attributes.sessionID)

	return response.Speak(welcomeMessage).Reprompt(helpMessage), nil
}

//  -----------------------
type repeatHandler struct{}

func (h *repeatHandler) CanHandle(input askgo.HandlerInput) bool {
	request := input.GetRequest()
	return request.Intent.Name == alexa.RepeatIntent
}
func (h *repeatHandler) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
	request := input.GetRequest()
	builder := input.GetResponse().WithShouldEndSession(false)
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	log.Printf("RepeatHandler requestId=%s, sessionId=%s", request.RequestID, attributes.sessionID)

	question := getQuestion(attributes)

	return builder.Speak(question).Reprompt(question), nil
}

//  -----------------------
type quizHandler struct{}

func (h *quizHandler) CanHandle(input askgo.HandlerInput) bool {
	request := input.GetRequest()

	return request.Intent.Name == "QuizIntent" || request.Intent.Name == alexa.StartOverIntent
}
func (h *quizHandler) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
	request := input.GetRequest()
	response := input.GetResponse().WithShouldEndSession(false)
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	log.Printf("QuizHandler requestId=%s, sessionId=%s", request.RequestID, attributes.sessionID)

	attributes.State = QUIZ
	attributes.Counter = 0
	askQuestion(request, attributes)
	question := getQuestion(attributes)

	if supportsDisplay(input.GetRequestEnvelope().Context) {
		title := fmt.Sprintf("Question #%v", attributes.Counter)

		image := &alexa.DisplayImageObject{}

		image.AddImageSource("", getBackgroundImage(getQuizItem(attributes).Abbreviation), 0, 0)

		itemList := make([]alexa.DisplayListItem, 0)
		for i, answer := range getMultipleChoiceAnswers(attributes) {
			itemList = append(itemList, alexa.DisplayListItem{
				Token: fmt.Sprintf("item_%d", i+1),
				TextContent: alexa.TextContent{
					PrimaryText: alexa.DisplayTextContent{
						Type: "PlainText",
						Text: answer,
					},
				},
			})
		}

		response.AddRenderTemplateDirective(alexa.DisplayTemplate{
			Type:  "ListTemplate1",
			Token: "QUESTION",
			// BackButton:      "HIDDEN",
			Title:           title,
			BackgroundImage: *image,
			/*
				TextContent: alexa.TextContent{
					PrimaryText: alexa.DisplayTextContent{
						Type: "RichText",
						Text: getQuestionWithoutOrdinal(attributes),
					},
				},
			*/
			ListItems: itemList,
		})
	}

	return response.Speak(fmt.Sprintf("%s %s", startQuizMessage, question)).Reprompt(question), nil
}

//  -----------------------
type definitionHandler struct{}

func (h *definitionHandler) CanHandle(input askgo.HandlerInput) bool {
	request := input.GetRequest()
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	return attributes.State != QUIZ && request.Intent.Name == "AnswerIntent"
}

func (h *definitionHandler) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
	request := input.GetRequest()
	response := input.GetResponse().WithShouldEndSession(false)
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	log.Printf("DefinitionHandler requestId=%s, sessionId=%s", request.RequestID, attributes.sessionID)

	overlap := make(map[string]int)

	var slotItem string
	for k, v := range request.Intent.Slots {
		if v.Value != "" {
			overlap[k] = 1
			slotItem = k
		}
	}

	s := structs.New(&data[0])
	for _, n := range s.Names() {
		if _, found := overlap[n]; found {
			overlap[n]++
		} else {
			overlap[n] = 1
		}
	}

	keys := make([]string, 0)
	for k, v := range overlap {
		if v == 2 {
			keys = append(keys, k)
		}
	}

	var match *QuizItem

	if len(keys) != 0 {
		key := keys[0]

		if item, ok := request.Intent.Slots[key]; ok {
			for _, entry := range data {
				s := structs.New(entry)
				v := s.Field(key).Value()
				if strings.EqualFold(fmt.Sprintf("%v", v), fmt.Sprintf("%v", item.Value)) {
					match = &entry
					break
				}
			}
		}
	}

	if match != nil {
		msg := getSpeechDescription(*match)

		/*
			if supportsDisplay(request.Context) {
				const image = new Alexa.ImageHelper().addImageInstance(getLargeImage(item)).getImage();
				const title = getCardTitle(item);
				const primaryText = new Alexa.RichTextContentHelper().withPrimaryText(getTextDescription(item, "<br/>")).getTextContent();
				response.addRenderTemplateDirective({
					type: 'BodyTemplate2',
					backButton: 'visible',
					image,
					title,
					textContent: primaryText,
				});
			}
		*/

		response.Speak(msg).Reprompt(msg)
	} else {
		msg := fmt.Sprintf("I'm sorry. %s is not something I know very much about in this skill. %s", formatCasing(slotItem), helpMessage)

		response.Speak(msg).Reprompt(msg)
	}

	return response, nil
}

//  -----------------------
type quizAnswerHandler struct{}

func (h *quizAnswerHandler) CanHandle(input askgo.HandlerInput) bool {
	request := input.GetRequest()
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	return attributes.State == QUIZ && request.Intent.Name == "AnswerIntent"
}
func (h *quizAnswerHandler) Handle(input askgo.HandlerInput) (*askgo.ResponseEnvelope, error) {
	request := input.GetRequest()
	response := input.GetResponse().WithShouldEndSession(false)
	attributes := input.GetContext().Value(&attributeContext).(*Attributes)

	log.Printf("QuizAnswerHandler requestId=%s, sessionId=%s", request.RequestID, attributes.sessionID)

	var isCorrect bool

	// Alexa is not good at putting the answer in the matching slot
	for _, prop := range request.Intent.Slots {
		if strings.EqualFold(prop.Value, attributes.QuizAnswer) {
			isCorrect = true
			break
		}
	}

	var cons string

	if isCorrect {
		attributes.QuizScore++

		cons = speechConsCorrect[random.Intn(len(speechConsCorrect))]
	} else {
		cons = speechConsWrong[random.Intn(len(speechConsWrong))]
	}

	output := []string{fmt.Sprintf("<say-as interpret-as='interjection'>%s</say-as><break strength='strong'/>", cons)}

	if attributes.Counter < 10 {
		askQuestion(request, attributes)
		question := getQuestion(attributes)

		output = append(output, question)
		response.Reprompt(question)

		/*
			if (supportsDisplay(handlerInput)) {
				const title = `Question #${attributes.counter}`;
				const primaryText = new Alexa.RichTextContentHelper().withPrimaryText(getQuestionWithoutOrdinal(attributes.quizProperty, attributes.quizItem)).getTextContent();
				const backgroundImage = new Alexa.ImageHelper().addImageInstance(getBackgroundImage(attributes.quizItem.Abbreviation)).getImage();
				const itemList = [];
				getAndShuffleMultipleChoiceAnswers(attributes.selectedItemIndex, attributes.quizItem, attributes.quizProperty).forEach((x, i) => {
					itemList.push(
						{
							"token" : x,
							"textContent" : new Alexa.PlainTextContentHelper().withPrimaryText(x).getTextContent(),
						}
					);
				});
				response.addRenderTemplateDirective({
					type : 'ListTemplate1',
					token : 'Question',
					backButton : 'hidden',
					backgroundImage,
					title,
					listItems : itemList,
				});
			}
		*/
	} else {
		output = append(output, getFinalScore(attributes))
		output = append(output, exitSkillMessage)

		/*
			if(supportsDisplay(handlerInput)) {
				const title = 'Thank you for playing';
				const primaryText = new Alexa.RichTextContentHelper().withPrimaryText(getFinalScore(attributes.quizScore, attributes.counter)).getTextContent();
				response.addRenderTemplateDirective({
					type : 'BodyTemplate1',
					backButton: 'hidden',
					title,
					textContent: primaryText,
				});
			}
		*/

		attributes.State = START
	}

	response.Speak(strings.Join(output, " "))

	return response, nil
}
