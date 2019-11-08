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

type Config struct {
	CsvUrl     string `json:"CSV_URL"`
	WebHookUrl string `json:"WEBHOOK_URL"`
	Channel    string `json:"CHANNEL"`
}

func run(config Config) {
	csvUrl:= config.get( "CSV_URL")
	data := readCSVFromUrl(csvUrl)

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
		webHookUrl := config.get("WEBHOOK_URL")
		channel := config.get("CHANNEL")
		text := ""
		for urlError := range failed {
			text += "`" + urlError + "`\n"
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

func (config Config) get(key string) string{
	var result string
	switch key {
	case "CSV_URL":
		result = config.CsvUrl
		break
	case "WEBHOOK_URL":
		result = config.WebHookUrl
		break
	case "CHANNEL":
		result = config.Channel
		break
	}
	if len(result) > 0 {
		return result
	}
	result, found := os.LookupEnv(key)
	if found {
		return result
	}

	panic(errors.New(key + " not found in env or json input"))
}

func readCSVFromUrl(url string) ([][]string) {
	resp, err := http.Get(url)
	if err != nil {
		panic(err)
	}
	defer resp.Body.Close()
	reader := csv.NewReader(resp.Body)
	reader.Comma = ';'
	data, err := reader.ReadAll()
	if err != nil {
		panic(err)
	}
	return data
}

func main() {
	lambda.Start(run)
}
