package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
)

const (
	version = "0.0.2"
	region  = "ap-northeast-1"
	iconURL = "https://ucarecdn.com/c847cfef-c658-43c3-9b4a-e3d4abf8534b/"
)

var (
	// USERNAME username of slack
	USERNAME = "GuardDutyAlert"
	// SLACKURL webhookurl of slack
	SLACKURL         = os.Getenv("SLACKURL")
	notPostThreshold = os.Getenv("THRESHOLD")
	config           = aws.Config{Region: aws.String(region)}
	svcIAM           = iam.New(session.New(&config))
)

// GuardDutyFindings set guardduty GuardDutyFindingsValue
type GuardDutyFindings struct {
	AccountID   string      `json:"accountId"`
	Region      string      `json:"region"`
	Type        string      `json:"type"`
	Severity    json.Number `json:"severity"`
	Title       string      `json:"title"`
	Description string      `json:"description"`
	Resource    Resource    `json:"resource"`
}

// Resource set guardduty ResourceValue
type Resource struct {
	ResourceType     string           `json:"resourceType,omitempty"`
	UserName         string           `json:"userName,omitempty"`
	InstanceDetails  InstanceDetails  `json:"instanceDetails,,omitempty"`
	AccessKeyDetails AccessKeyDetails `json:"accessKeyDetails,,omitempty"`
}

// InstanceDetails set guardduty value
type InstanceDetails struct {
	InstanceID   string `json:"instanceId,omitempty"`
	InstanceType string `json:"instanceType,"`
}

// AccessKeyDetails set guardduty AccessKeyDetailsValue
type AccessKeyDetails struct {
	UserName string `json:"userName,omitempty"`
}

func main() {
	lambda.Start(Handler)
}

// Handler get value from cloudwatch event
func Handler(event events.CloudWatchEvent) (events.CloudWatchEvent, error) {
	var resource string
	gd := &GuardDutyFindings{}

	err := json.Unmarshal([]byte(event.Detail), gd)
	fmt.Println(string([]byte(event.Detail)))
	if err != nil {
		fmt.Println(err)
	}

	// cast to float64
	float64Severity, err := gd.Severity.Float64()
	slackColor := CheckSeverityLevel(float64Severity)
	float64NotPostThreshold, err := strconv.ParseFloat(notPostThreshold, 64)

	// get aws account name
	accountAliasName := FetchAccountAlias()

	// Set affected resource
	if gd.Resource.InstanceDetails.InstanceID != "" {
		resource = gd.Resource.InstanceDetails.InstanceID
	} else if gd.Resource.AccessKeyDetails.UserName != "" {
		resource = gd.Resource.AccessKeyDetails.UserName
	} else {
		resource = "unknown"
	}

	// Post slack
	if float64Severity > float64NotPostThreshold {
		PostSlack(slackColor, gd.Title, accountAliasName, string(gd.Severity), resource, gd.Type, gd.Description)
	}
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
func PostSlack(slackColor string, title string, accountAliasName string, severity string, resource string, reason string, description string) {
	field0 := slack.Field{Title: "Title", Value: "_" + title + "_"}
	field1 := slack.Field{Title: "Account", Value: accountAliasName}
	field2 := slack.Field{Title: "Severity", Value: severity}
	field3 := slack.Field{Title: "Affected Resource", Value: resource}
	field4 := slack.Field{Title: "Type", Value: reason}
	field5 := slack.Field{Title: "Description", Value: "```" + description + "```"}

	attachment := slack.Attachment{}
	attachment.AddField(field0).AddField(field1).AddField(field2).AddField(field3).AddField(field4).AddField(field5)
	color := slackColor
	attachment.Color = &color
	payload := slack.Payload{
		Username:    USERNAME,
		IconUrl: iconURL,
		Attachments: []slack.Attachment{attachment},
	}
	err := slack.Send(SLACKURL, "", payload)
	if err != nil {
		os.Exit(1)
	}
}
