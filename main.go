package main

import (
	"encoding/csv"
	"errors"
	"fmt"
	"github.com/ashwanthkumar/slack-go-webhook"
	"github.com/aws/aws-lambda-go/lambda"
	"net/http"
	"os"
	"strings"
	"sync"
)

type JsonInputEvent struct {
	CsvUrl     string `json:"CSV_URL"`
	WebHookUrl string `json:"WEBHOOK_URL"`
	Channel    string `json:"CHANNEL"`
}

func run(jsonInput JsonInputEvent) {
	csvUrl, err := getCsvUrl(jsonInput)
	if err != nil {
		panic(err)
	}
	data, err := readCSVFromUrl(csvUrl)
	if err != nil {
		panic(err)
	}

	var waitGroup sync.WaitGroup
	failed := make(chan string, len(data))
	successes := make(chan string, len(data))
	for _, row := range data {
		// skip header or invalid row
		if strings.Index(row[0], "http") == -1 {
			continue
		}
		waitGroup.Add(1)
		go checkUrl(&waitGroup, row[0], failed, successes)
	}

	waitGroup.Wait()
	close(failed)
	close(successes)

	if len(failed) > 0 {
		webHookUrl, err := getWebHookUrl(jsonInput)
		if err != nil {
			panic(err)
		}

		channel, err := getChannel(jsonInput)
		if err != nil {
			panic(err)
		}
		var text string = ""

		for urlError := range failed {
			text += fmt.Sprintf("`" + urlError + "`\n")
		}
		payload := slack.Payload{
			Text:      ":warning:Errors found in the following domains :\n" + text,
			Username:  "robot",
			Channel:   "#"+channel,
			IconEmoji: ":warning:",
		}
		slackError := slack.Send(webHookUrl, "", payload)
		if slackError != nil {
			fmt.Printf("error: %s\n", slackError)
		}
	}
	fmt.Printf("Successfully checked %d", len(successes))
}

func checkUrl(waitGroup *sync.WaitGroup, url string, failed chan<- string, successes chan<- string) {
	defer waitGroup.Done()
	req, err := http.Get(url)
	if err != nil {
		failed <- url + " " + err.Error()
	} else {
		successes <- url + " " + req.Status
	}
}

func getCsvUrl(jsonInput JsonInputEvent) (string, error) {
	csvUrl, found := os.LookupEnv("CSV_URL")
	if !found && len(csvUrl) == 0 && len(jsonInput.CsvUrl) == 0 {
		return "", errors.New("CSV_URL not found in env or json input")
	}
	if len(jsonInput.CsvUrl) > 0 {
		return jsonInput.CsvUrl, nil
	} else {
		return csvUrl, nil
	}
}

func getWebHookUrl(jsonInput JsonInputEvent) (string, error) {
	webHookUrl, found := os.LookupEnv("WEBHOOK_URL")
	if !found && len(webHookUrl) == 0 && len(jsonInput.WebHookUrl) == 0 {
		return "", errors.New("WEBHOOK_URL not found in env or json input")
	}
	if len(jsonInput.WebHookUrl) > 0 {
		return jsonInput.WebHookUrl, nil
	} else {
		return webHookUrl, nil
	}
}

func getChannel(jsonInput JsonInputEvent) (string, error) {
	channel, found := os.LookupEnv("CHANNEL")
	if !found && len(channel) == 0 && len(jsonInput.Channel) == 0 {
		return "", errors.New("CHANNEL not found in env or json input")
	}
	if len(jsonInput.Channel) > 0 {
		return jsonInput.Channel, nil
	} else {
		return channel, nil
	}
}

func readCSVFromUrl(url string) ([][]string, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	data, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}
	return data, nil
}

func main() {
	lambda.Start(run)
}
