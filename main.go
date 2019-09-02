package main

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const version = "0.0.1"

// GuardDutyFindings set value from json
type GuardDutyFindings struct {
	Title       string      `json:"title"`
	Type        string      `json:"type"`
	AccountID   string      `json:"accountID"`
	Description string      `json:"description"`
	Severity    json.Number `json:"severity"`
}

var (
	// USERNAME username of slack
	USERNAME = "GuardDutyAlert"

	// SLACKURL webhookurl of slack
	SLACKURL = os.Getenv("SLACKURL")

	config = aws.Config{Region: aws.String("ap-northeast-1")}
	svcIAM = iam.New(session.New(&config))
)

func main() {
	lambda.Start(Handler)
}

// Handler get value from cloudwatch event
func Handler(event events.CloudWatchEvent) (events.CloudWatchEvent, error) {
	gd := &GuardDutyFindings{}

	err := json.Unmarshal([]byte(event.Detail), gd)
	if err != nil {
		fmt.Println(err)
	}

	// cast to float54
	float64Severity, err := gd.Severity.Float64()
	slackColor := CheckSeverityLevel(float64Severity)

	// get aws account name
	accountAliasName := FetchAccountAlias()

	// post slack
	PostSlack(slackColor, accountAliasName, string(gd.Severity), gd.Type, gd.Description)

	return event, err
}

// CheckSeverityLevel fix the color
func CheckSeverityLevel(severity float64) string {
	var color string

	if severity == 0.0 {
		color = "good"
	} else if (0.1 <= severity) && (severity <= 3.9) {
		color = "#0000ff"
	} else if (4.0 <= severity) && (severity <= 6.9) {
		color = "warning"
	} else {
		color = "danger"
	}
	return color
}

// FetchAccountAlias get account alias name
func FetchAccountAlias() string {
	var accountAlias string

	params := &iam.ListAccountAliasesInput{}
	res, err := svcIAM.ListAccountAliases(params)
	if err != nil {
		fmt.Println(err)
	}
	if res.AccountAliases == nil {
		accountAlias = "None"
	} else {
		accountAlias = *res.AccountAliases[0]
	}
	return accountAlias
}

// PostSlack post slack result
func PostSlack(slackColor string, accountAliasName string, severity string, reason string, description string) {
	field1 := slack.Field{Title: "Account", Value: accountAliasName}
	field2 := slack.Field{Title: "Severity", Value: severity}
	field3 := slack.Field{Title: "Type", Value: reason}
	field4 := slack.Field{Title: "Description", Value: "```" + description + "```"}

	attachment := slack.Attachment{}
	attachment.AddField(field1).AddField(field2).AddField(field3).AddField(field4)
	color := slackColor
	attachment.Color = &color
	payload := slack.Payload{
		Username:    USERNAME,
		Attachments: []slack.Attachment{attachment},
	}
	err := slack.Send(SLACKURL, "", payload)
	if err != nil {
		fmt.Println(err)
	}
}
