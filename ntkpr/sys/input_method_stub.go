//go:build !darwin
// +build !darwin

package sys

import "errors"

type InputMethodType string

const (
	InputMethodEnglish InputMethodType = "en"
	InputMethodIME     InputMethodType = "ime"
	InputMethodUnknown InputMethodType = "unknown"
)

var ErrUnsupported = errors.New("input method not supported on this OS")

func GetCurrentInputMethod() (InputMethodType, string) {
	return InputMethodUnknown, ""
}

func SwitchInputMethod(id string) error {
	return ErrUnsupported
}

func InputMethodID(t InputMethodType) (string, error) {
	return "", ErrUnsupported
}
