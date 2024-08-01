package google

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
)

type GoogleError struct {
	Domain       string `json:"domain"`
	Reason       string `json:"reason"`
	Message      string `json:"message"`
	LocationType string `json:"locationType"`
	Location     string `json:"location"`
}

func NewGoogleError(domain, reason, message, locationType, location string) *GoogleError {
	return &GoogleError{Domain: domain, Reason: reason, Message: message, LocationType: locationType, Location: location}
}

func (e *GoogleError) Error() string {
	return e.Message
}

// Is implements the [errors.Is].
//
// The Domain, Reason, LocationType and the Location fields must be equal.
// If the Message field is not equal, than the Message field of target
// can be used used as a regexp pattern to allow matching errors with dynamic fields (eg.: ErrLighthouseInvalidUrl)
func (e *GoogleError) Is(target error) bool {

	t, ok := target.(*GoogleError)
	if !ok {
		// Not *GoogleError
		return false
	}

	// Domain differs
	if t.Domain != e.Domain {
		return false
	}

	// Reason differs
	if t.Reason != e.Reason {
		return false
	}

	// LocationType differs
	if t.LocationType != e.LocationType {
		return false
	}

	// Location differs
	if t.Location != e.Location {
		return false
	}

	// Message differs
	if t.Message != e.Message {

		// But Message can be used as a regexp pattern
		r, _ := regexp.MatchString(t.Message, e.Message)

		return r
	}

	return true
}

// Error stores the Standard Error Messages.
//
// See: https://developers.google.com/webmaster-tools/v1/errors
type Error struct {
	Code    int     `json:"code"`
	Message string  `json:"message"`
	Errors  []error `json:"errors"`

	data []byte
}

func NewError(code int, message string) *Error {
	return &Error{Code: code, Message: message}
}

// ErrorFromResponse reads the response body and returns the Error.
//
// The response data is stored in Error and returned by [String]
func ErrorFromResponse(r *http.Response) (*Error, error) {

	data, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, fmt.Errorf("read error: %w", err)
	}

	v := struct {
		Err *Error `json:"error"`
	}{}

	err = json.Unmarshal(data, &v)
	if err != nil {
		return nil, fmt.Errorf("unmarshal error: %w", err)
	}

	if v.Err != nil {
		v.Err.data = data
	}

	return v.Err, nil
}

func (e *Error) Unwrap() []error {
	return e.Errors
}

// String returns the stored JSON data or (if data is nil) returns e.Message.
func (e *Error) String() string {
	if len(e.data) == 0 {
		return e.Message
	}
	return string(e.data)
}

// Error returns the e.Message.
func (e *Error) Error() string {
	return e.Message
}

func (e *Error) UnmarshalJSON(data []byte) error {

	v := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Errors  []struct {
			Domain       string `json:"domain"`
			Reason       string `json:"reason"`
			Message      string `json:"message"`
			LocationType string `json:"locationType"`
			Location     string `json:"location"`
		} `json:"errors"`
	}{}

	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	r := new(Error)

	r.Code = v.Code
	r.Message = v.Message

	for i := range v.Errors {
		r.Errors = append(r.Errors, NewGoogleError(v.Errors[i].Domain, v.Errors[i].Reason, v.Errors[i].Message, v.Errors[i].LocationType, v.Errors[i].Location))
	}

	*e = *(*Error)(r)

	return nil
}
