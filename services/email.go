package services

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"

	"github.com/Origho-precious/url-shortener/go/configs"
)

const (
	method = "POST"
	url    = "https://send.api.mailtrap.io/api/send"
)

type Email struct {
	TemplateUUID string
	Title        string
	Recipients   []string
	Extra        map[string]string
}

type Recipient struct {
	Email string `json:"email"`
}

type Payload struct {
	To             []Recipient       `json:"to"`
	From           map[string]string `json:"from"`
	TemplateUUID   string            `json:"template_uuid"`
	TemplateExtras map[string]string `json:"template_variables"`
}

func (em Email) SendEmail() (string, error) {
	cfg, err := configs.LoadEnvs()
	if err != nil {
		return "", err
	}

	fromEmail := cfg.MAILTRAP_SENDER_EMAIL

	var recipients []Recipient
	for _, email := range em.Recipients {
		recipients = append(recipients, Recipient{Email: email})
	}

	data := Payload{
		From: map[string]string{
			"name":  em.Title,
			"email": fromEmail,
		},
		TemplateExtras: em.Extra,
		To:             recipients,
		TemplateUUID:   em.TemplateUUID,
	}

	payload, err := json.Marshal(data)
	if err != nil {
		return "", err
	}

	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewBuffer(payload))
	if err != nil {
		return "", err
	}

	req.Header.Add("Authorization", cfg.MAILTRAP_AUTH)
	req.Header.Add("Content-Type", "application/json")

	res, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	return string(body), nil
}
