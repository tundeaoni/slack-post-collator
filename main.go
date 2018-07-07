package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

const PORT = "8080"
const VERIFICATION_TYPE = "url_verification"
const EVENT_TYPE = "event_callback"
const REACTION_REMOVE_TYPE = "reaction_removed"
const REACTION_ADDED_TYPE = "reaction_added"
const SLACK_TOKEN = ""
const SLACK_HISTORY_END_POINT = "https://slack.com/api/channels.history"

type Envelope struct {
	Type string
}

type Verification struct {
	Token     string
	Challenge string
	Type      string
}

type EventItem struct {
	EventTime int64 `json:"event_time"`
	Event     struct {
		Type     string
		User     string
		Reaction string
		ItemUser string `json:"item_user"`
		Item     struct {
			Type    string
			Channel string
			Ts      string
		}
	}
}

type Messages struct {
	Messages []Message
	HasMore  bool `json:"has_more"`
}

type Message struct {
	User string
	Text string
	Ts   string
}

var (
	envelope     Envelope
	verification Verification
	item         EventItem
	messages     Messages
	err          error
)

func getMessage(timeStamp, channel string) Messages {
	response, err := http.Get(SLACK_HISTORY_END_POINT + "?inclusive=1&count=1" + "&channel=" + channel + "&token=" + SLACK_TOKEN + "&latest=" + timeStamp + "&oldest=" + timeStamp)
	if err != nil {
		panic(err)
	}
	defer response.Body.Close()
	content, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(content, &messages)
	if err != nil {
		panic(err)
	}
	return messages
}

func process(w http.ResponseWriter, req *http.Request) {
	//  http://eagain.net/articles/go-dynamic-json/
	body, readError := ioutil.ReadAll(req.Body)
	if readError != nil {
		panic(readError)
	}

	err = json.Unmarshal(body, &envelope)
	if readError != nil {
		panic(err)
	}
	defer req.Body.Close()

	payloadType := envelope.Type
	if payloadType == VERIFICATION_TYPE {
		err = json.Unmarshal(body, &verification)
		if readError != nil {
			panic(err)
		}

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(verification.Challenge))

	} else if payloadType == EVENT_TYPE {
		// TODO: filter based on channel
		// TODO: filter based on reaction

		err = json.Unmarshal(body, &item)
		if readError != nil {
			panic(err)
		}

		result := getMessage(item.Event.Item.Ts, item.Event.Item.Channel)
		if len(result.Messages) == 0 {
			w.WriteHeader(http.StatusOK)
			return
		}

		message := result.Messages[0]

		switch eventType := item.Event.Type; eventType {
		case REACTION_REMOVE_TYPE:
			fmt.Println("removed reaction on message: " + message.Text)
		case REACTION_ADDED_TYPE:
			fmt.Println("added reaction on message: " + message.Text)
		default:

		}
	}

}

func main() {
	http.HandleFunc("/", process)

	fmt.Println("starting app on " + PORT)
	if err := http.ListenAndServe(":"+PORT, nil); err != nil {
		panic(err)
	}

}
