package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"time"
)

// SlackSender has the hook to send slack notifications.
type SlackSender struct {
	Hook string
}

type AttachmentField struct {
  Title string `json:"title"` 
  Value string `json:"value"`
  Short bool   `json:"short"`
}

type SlackAttachment struct {
  Fallback string `json:"fallback"`
  Text string `json:"text"`
  Pretext  string `json:"pretext"`
  Color    string `json:"color"`
  Title string `json:"title"`
  TitleLink string `json:"title_link"`
  Fields   []AttachmentField `json:"fields"`
  Footer string `json:"footer"`
  FooterIcon string `json:"footer_icon"`
  MarkdownIn []string `json:"mrkdwn_in"`
  TS int64 `json:"ts"`
}

type SlackPayload struct {
  Attachments []SlackAttachment `json:"attachments"`
}

// Send a notification with a formatted message build from the repository.
func (s *SlackSender) Send(repository Repository) error {
  var borderColor string;
  if (repository.Release.IsPrerelease) {
    borderColor = "#FFC600";
  } else {
    borderColor = "#15e415";
  }

  textContent := fmt.Sprintf(
    "<%s|%s/%s>: <%s|%s> released",
    repository.URL.String(),
    repository.Owner,
    repository.Name,
    repository.Release.URL.String(),
    repository.Release.Name,
  );

  payload := SlackPayload{
    Attachments: []SlackAttachment{{
      Fallback: textContent,
      Text: textContent,
      Pretext: fmt.Sprintf(
        "*%s/%s* - _%s_",
        repository.Owner,
        repository.Name,
        repository.Release.Name,
      ),
      Title: repository.Release.Name,
      TitleLink: repository.Release.URL.String(),
      Color: borderColor,
      Fields: []AttachmentField{
        {
          Title: "Description",
          Value: repository.Release.Description,
          Short: false,
        },
      },
      TS: repository.Release.PublishedAt.Unix(),
      Footer: " ",
      FooterIcon: "https://assets-cdn.github.com/images/modules/logos_page/GitHub-Mark.png",
      MarkdownIn: []string{"pretext"},
    }},
  }

	payloadData, err := json.Marshal(payload)
	if err != nil {
		return err
	}

	req, err := http.NewRequest(http.MethodPost, s.Hook, bytes.NewReader(payloadData))
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	req = req.WithContext(ctx)
	defer cancel()

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := ioutil.ReadAll(resp.Body)
		return fmt.Errorf("request didn't respond with 200 OK: %s, %s", resp.Status, body)
	}

	return nil
}
