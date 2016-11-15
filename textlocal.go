package textlocal

import (
	"errors"
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strings"
	"net/url"
)

// DefaultBaseUrl is the default TextLocal API gateway.
const DefaultBaseUrl = "https://api.txtlocal.com"

// ErrEmptyBaseUrl empty base URL can not be set.
var ErrEmptyBaseUrl = errors.New("You must set a base url. Use textlocal.DefaultBaseUrl to use default.")

// ErrNoApiKeyProvided no API key has been provided.
var ErrNoApiKeyProvided = errors.New("You must provide an API key.")

// ErrUnknown when an unknown error is returned.
var ErrUnknown = errors.New("API error, unknown.");

// ErrCommandNotValid when a command is not found on API endpoint.
var ErrCommandNotValid = errors.New("API error, unknown command.");

// ErrInvalidApiKey when a response comes back with API key error.
var ErrInvalidApiKey = errors.New("API error, invalid api key provided.");

const (
	_ = iota
	noCommandSpecifiedErrorCode
	unrecognisedCommandErrorCode
	invalidLoginDetailsErrorCode
)

func getErrorFromResponse(b map[string]interface{}) error {
	if e, ok := b["error"].([]map[string]interface{}); ok {
		if len(e) > 0 {
			te := e[0]

			switch te["code"].(int) {
			case noCommandSpecifiedErrorCode:
				return ErrCommandNotValid
			case unrecognisedCommandErrorCode:
				return ErrCommandNotValid
			case invalidLoginDetailsErrorCode:
				return ErrInvalidApiKey
			}
		}
	}

	return ErrUnknown
}

// GetCreditsResponse contains how many credits remain.
type GetCreditsResponse struct {
	RemainingSMS int
	RemainingMMS int
}

// Connection for connecting to TextLocal
type Connection struct {
	baseUrl string
	apiKey  string
	client  *http.Client
}

// New connection.
func New(baseUrl, apiKey string) (*Connection, error) {
	return NewWithCustomHttpClient(baseUrl, apiKey, http.DefaultClient)
}

// New connection.
func NewWithCustomHttpClient(baseUrl, apiKey string, client *http.Client) (*Connection, error) {
	if len(baseUrl) == 0 {
		return &Connection{}, ErrEmptyBaseUrl
	}

	if len(apiKey) == 0 {
		return &Connection{}, ErrNoApiKeyProvided
	}

	return &Connection{
		baseUrl: strings.TrimRight(baseUrl, "/"),
		apiKey:  apiKey,
		client:  client,
	}, nil
}

// GetBaseUrl returns the base url used to construct the connection.
func (c *Connection) GetBaseUrl() string {
	return c.baseUrl
}

// GetCredits will tell you amount of SMS and MMS credits available on the account.
func (c *Connection) GetCredits() (GetCreditsResponse, error) {
	res, err := c.client.Get(c.baseUrl + "/balance?apiKey=" + c.apiKey)

	if err != nil {
		return GetCreditsResponse{}, err
	}

	content, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return GetCreditsResponse{}, err
	}

	var payload map[string]interface{}

	if err = json.Unmarshal(content, &payload); err != nil {
		return GetCreditsResponse{}, err
	}

	if payload["status"].(string) != "success" {
		return GetCreditsResponse{}, getErrorFromResponse(payload)
	}

	rS, rM := 0, 0

	if ba, ok := payload["balance"].(map[string]interface{}); ok {
		rS = int(ba["sms"].(float64))
		rM = int(ba["mms"].(float64))
	}

	return GetCreditsResponse{
		RemainingSMS: rS,
		RemainingMMS: rM,
	}, nil
}

// SendSMS will send a message to a list of numbers.
func (c *Connection) SendSMS(n []string, m string, f string) (error) {
	form := url.Values{}
	form.Add("apiKey", c.apiKey)
	form.Add("numbers", strings.Join(n, ","))
	form.Add("message", m)
	form.Add("sender", f)

	req, err := http.NewRequest("POST", c.baseUrl + "/send", strings.NewReader(form.Encode()))
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	res, err := c.client.Do(req)

	if err != nil {
		return err
	}

	content, err := ioutil.ReadAll(res.Body)

	if err != nil {
		return err
	}

	var payload map[string]interface{}

	if err = json.Unmarshal(content, &payload); err != nil {
		return err
	}

	if payload["status"].(string) != "success" {
		return getErrorFromResponse(payload)
	}

	return nil
}
