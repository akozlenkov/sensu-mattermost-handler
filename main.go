package main

import (
	"fmt"
	"github.com/go-resty/resty/v2"
	"github.com/sensu/sensu-go/types"
	"github.com/sensu/sensu-plugin-sdk/sensu"
	"time"
)

type Field struct {
	Title string `json:"title"`
	Value string `json:"value"`
	Short bool   `json:"short"`
}

type Attachment struct {
	Fallback  string  `json:"fallback"`
	Color     string  `json:"color"`
	Pretext   string  `json:"pretext"`
	Title     string  `json:"title"`
	Text      string  `json:"text"`
	TitleLink string  `json:"title_link"`
	Fields    []Field `json:"fields"`
}

type Message struct {
	Channel     string       `json:"channel"`
	Username    string       `json:"username"`
	IconUrl     string       `json:"icon_url"`
	Text        string       `json:"text"`
	Attachments []Attachment `json:"attachments"`
}

type handlerConfig struct {
	sensu.PluginConfig
	endpoint string
	channel  string
	username string
	iconUrl  string
}

var (
	config = handlerConfig{
		PluginConfig: sensu.PluginConfig{
			Name:  "sensu-mattermost-handler",
			Short: "The Sensu Go mattermost handler",
		},
	}

	handlerOptions = []*sensu.PluginConfigOption{
		{
			Path:     "endpoint",
			Env:      "MATTERMOST_ENDPOINT",
			Argument: "endpoint",
			Usage:    "Mattermost endpoint",
			Value:    &config.endpoint,
		},
		{
			Path:     "channel",
			Env:      "MATTERMOST_CHANNEL",
			Argument: "channel",
			Usage:    "Mattermost channel",
			Value:    &config.channel,
		},
		{
			Path:     "username",
			Env:      "MATTERMOST_USERNAME",
			Argument: "username",
			Usage:    "Mattermost username",
			Default:  "sensu",
			Value:    &config.username,
		},
		{
			Path:     "icon_url",
			Env:      "MATTERMOST_ICON_URL",
			Argument: "icon_url",
			Usage:    "Mattermost icon url",
			Default:  "https://cdn.iconscout.com/icon/free/png-128/sensu-3627944-3029170.png",
			Value:    &config.iconUrl,
		},
	}
)

func main() {
	sensu.NewGoHandler(&config.PluginConfig, handlerOptions,
		func(event *types.Event) error {
			if len(config.endpoint) == 0 {
				return fmt.Errorf("--endpoint or MATTERMOST_ENDPOINT environment variable is required")
			}
			if len(config.channel) == 0 {
				return fmt.Errorf("--channel or MATTERMOST_CHANNEL environment variable is required")
			}
			return nil
		},
		func(event *types.Event) error {
			attachment := Attachment{
				Title: event.Check.Name,
				Text:  event.Check.Output,
				Fields: []Field{
					{Title: "Entity", Value: event.Entity.Name, Short: true},
					{Title: "Entity class", Value: event.Entity.EntityClass, Short: true},
					{Title: "Status", Value: event.Check.State, Short: true},
					{Title: "Executed", Value: time.Unix(event.Check.Executed, 0).Format(time.RFC1123), Short: true},
				},
			}

			switch event.Check.Status {
			case 0:
				attachment.Color = "good"
			case 1:
				attachment.Color = "warning"
			case 2:
				attachment.Color = "danger"
			}

			if _, err := resty.New().R().SetBody(Message{
				Channel:     config.channel,
				Username:    config.username,
				IconUrl:     config.iconUrl,
				Attachments: []Attachment{attachment},
			}).Post(config.endpoint); err != nil {
				return err
			}
			return nil
		},
	).Execute()
}
