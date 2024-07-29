package google

import (
	"encoding/json"
	"fmt"
)

// RuntimeError stores the response code and the error message.
type RuntimeError struct {
	Code    int   `json:"code"`
	Message error `json:"message"`
}

func NewRuntimeError(code int, message error) RuntimeError {
	return RuntimeError{Code: code, Message: message}
}

func (r RuntimeError) Error() string {
	return fmt.Sprintf("%d %s", r.Code, r.Message)
}

func (r *RuntimeError) UnmarshalJSON(data []byte) error {

	v := struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
	}{}

	err := json.Unmarshal(data, &v)
	if err != nil {
		return err
	}

	// *r = *(*RuntimeError)(&v)
	r.Code = v.Code
	r.Message = fmt.Errorf(v.Message)

	return err
}
