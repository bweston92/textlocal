package textlocal

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/jarcoal/httpmock"
)

func TestGetErrorFromResponse_ErrUnknown(t *testing.T) {
	assert.Equal(t, ErrUnknown, getErrorFromResponse(map[string]interface{}{}))
}

func TestGetErrorFromResponse_ErrCommandNotValid_FromNoCommandSpecified(t *testing.T) {
	assert.Equal(t, ErrCommandNotValid, getErrorFromResponse(map[string]interface{}{
		"error": []map[string]interface{}{
			map[string]interface{}{
				"code": noCommandSpecifiedErrorCode,
				"message": "No command specified.",
			},
		},
	}))
}

func TestGetErrorFromResponse_ErrCommandNotValid_FromUnrecognisedCommand(t *testing.T) {
	assert.Equal(t, ErrCommandNotValid, getErrorFromResponse(map[string]interface{}{
		"error": []map[string]interface{}{
			map[string]interface{}{
				"code": unrecognisedCommandErrorCode,
				"message": "Unrecognised command.",
			},
		},
	}))
}

func TestGetErrorFromResponse_ErrInvalidApikey(t *testing.T) {
	assert.Equal(t, ErrInvalidApiKey, getErrorFromResponse(map[string]interface{}{
		"error": []map[string]interface{}{
			map[string]interface{}{
				"code": invalidLoginDetailsErrorCode,
				"message": "Invalid login details",
			},
		},
	}))
}

func TestThatAnErrorIsReturnedIfNoBaseUrlIsSet(t *testing.T) {
	_, e := New("", "api-key")

	assert.Equal(t, ErrEmptyBaseUrl, e)
}

func TestThatAnErrorIsReturnedIfApiKeyIsSet(t *testing.T) {
	_, e := New(DefaultBaseUrl, "")

	assert.Equal(t, ErrNoApiKeyProvided, e)
}

func TestItTrimsSlashFromTheEndOfBaseUrl(t *testing.T) {
	c, _ := New("http://localhost/", "api-key")

	assert.Equal(t, "http://localhost", c.GetBaseUrl())
}

func TestCreditsCanBeObtained(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	c, _ := New("https://api.txtlocal.com", "abc")

	res := httpmock.NewStringResponder(200, `{"balance": {"sms": 361.5, "mms": 0}, "status": "success"}`)
	httpmock.RegisterResponder("GET", "https://api.txtlocal.com/balance?apiKey=abc", res)

	cr, err := c.GetCredits()

	assert.Nil(t, err)
	assert.Equal(t, 361, cr.RemainingSMS)
	assert.Equal(t, 0, cr.RemainingMMS)
}

func TestCreditsCanBeFailQuietly(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	c, _ := New("https://api.txtlocal.com", "abc")

	res := httpmock.NewStringResponder(200, `{"status": "failure"}`)
	httpmock.RegisterResponder("GET", "https://api.txtlocal.com/balance?apiKey=abc", res)

	_, err := c.GetCredits()

	assert.NotNil(t, err)
}
