package google_test

import (
	"errors"
	"testing"

	"github.com/g0rbe/go-google"
)

func TestGoogleErrorIs(t *testing.T) {
	if !errors.Is(google.ErrLighthouseFailedDocumentRequest, google.ErrLighthouseFailedDocumentRequest) {
		t.Fatalf("ErrLighthouseFailedDocumentRequest is NOT ErrLighthouseFailedDocumentRequest\n")
	}
	if !errors.Is(google.ErrLighthouseRateLimitExceeded, google.ErrLighthouseRateLimitExceeded) {
		t.Fatalf("ErrLighthouseRateLimitExceeded is NOT ErrLighthouseRateLimitExceeded\n")
	}

}

func TestErrorIs(t *testing.T) {

	gErr := google.NewGoogleError("global", "invalidParameter", "Invalid string value: 'asdf'. Allowed values: [mostpopular]", "", "")

	err := &google.Error{
		Code:    400,
		Message: "Invalid string value: 'asdf'. Allowed values: [mostpopular]",
		Errors: []error{
			google.NewGoogleError("global", "invalidParameter", "Invalid string value: 'asdf'. Allowed values: [mostpopular]", "", ""),
		},
	}

	if !errors.Is(err, gErr) {
		t.Fatalf("FAIL: not equal\n")
	}
}

func TestErrorIsErrLighthouseUnprocessable(t *testing.T) {

	err := &google.Error{
		Code:    400,
		Message: "Invalid string value: 'asdf'. Allowed values: [mostpopular]",
		Errors: []error{
			google.NewGoogleError("global", "internalError", "Unable to process request. Please wait a while and try again.", "", ""),
			//google.ErrLighthouseUnprocessable,
		},
	}

	if !errors.Is(err, google.ErrLighthouseUnprocessable) {
		t.Fatalf("FAIL: error is not ErrLighthouseUnprocessable\n")
	}
}

func TestLighthouseErrorIsErrLighthouseUnprocessable(t *testing.T) {

	err := &google.LighthouseError{
		URL: "https://example.com",
		Err: &google.Error{
			Code:    400,
			Message: "Invalid string value: 'asdf'. Allowed values: [mostpopular]",
			Errors: []error{
				&google.GoogleError{
					Domain:  "global",
					Reason:  "internalError",
					Message: "Unable to process request. Please wait a while and try again.",
				},
			},
		},
	}

	if !errors.Is(err, google.ErrLighthouseUnprocessable) {
		t.Fatalf("FAIL: error is not ErrLighthouseUnprocessable\n")
	}
}
