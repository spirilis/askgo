package main

import (
	"context"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/spirilis/askgo"
)

func main() {
	skill := &askgo.Skill{
		// Uncomment the ApplicationID to prevent someone else from configuring a skill that sends requests to this function.
		// ApplicationID: "xyzzy",

		// Timestamps are not important for this skill, plus aids in debugging
		IgnoreTimestamp: true,

		RequestInterceptors: []askgo.RequestInterceptor{
			&unpackAttributes{},
		},

		// Our handler interface
		Handlers: []askgo.RequestHandler{
			&sessionEndHandler{},
			&launchHandler{},
			&exitHandler{},
			&helpHandler{},
			&repeatHandler{},
			&quizHandler{},
			&definitionHandler{},
			&quizAnswerHandler{},
		},

		ResponseInterceptors: []askgo.ResponseInterceptor{
			&saveAttributes{},
		},
	}

	lambda.Start(func(ctx context.Context, envelope *askgo.RequestEnvelope) (interface{}, error) {
		return skill.ProcessRequest(askgo.NewDefaultHandler(ctx, envelope))
	})
}
