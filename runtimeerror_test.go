package google_test

import (
	"encoding/json"
	"testing"

	"github.com/g0rbe/go-google"
)

func TestUnmarshalRunetimeError(t *testing.T) {
	resp := []byte(`
{
  "error": {
    "code": 400,
    "message": "Lighthouse returned error: FAILED_DOCUMENT_REQUEST. Lighthouse was unable to reliably load the page you requested. Make sure you are testing the correct URL and that the server is properly responding to all requests. (Details: net::ERR_TIMED_OUT)",
    "errors": [
      {
        "message": "Lighthouse returned error: FAILED_DOCUMENT_REQUEST. Lighthouse was unable to reliably load the page you requested. Make sure you are testing the correct URL and that the server is properly responding to all requests. (Details: net::ERR_TIMED_OUT)",
        "domain": "lighthouse",
        "reason": "lighthouseUserError"
      }
    ]
  }
}
`)

	v := struct {
		Err google.RuntimeError `json:"error"`
	}{}

	err := json.Unmarshal(resp, &v)
	if err != nil {
		t.Fatalf("Failed to unmarshal: %s\n", err)
	}

	if v.Err.Code != 400 {
		t.Fatalf("Invalid code: want: 400, got %d\n", v.Err.Code)
	}
}
