package sys

/*
#cgo LDFLAGS: -framework Carbon
#include <Carbon/Carbon.h>
*/
import "C"
import (
	"errors"
	"strings"
	"unsafe"
)

type InputMethodType string

const (
	InputMethodEnglish InputMethodType = "en"
	InputMethodIME     InputMethodType = "ime" // Chinese / Japanese / Korean
	InputMethodUnknown InputMethodType = "unknown"
)

// these should also be in configs. Or auto detection and load.
var ENGLISH_INPUT_METHOD_ID string = "com.apple.keylayout.US"
var CHINESE_INPUT_METHOD_ID string = "com.apple.inputmethod.SCIM.ITABC"

// GetCurrentInputMethod returns (input method type, rawID)
func GetCurrentInputMethod() (InputMethodType, string) {
	source := C.TISCopyCurrentKeyboardInputSource()
	if source == 0 {
		return InputMethodUnknown, ""
	}
	defer C.CFRelease(C.CFTypeRef(source))

	idPtr := C.TISGetInputSourceProperty(
		source,
		C.kTISPropertyInputSourceID,
	)
	if idPtr == nil {
		return InputMethodUnknown, ""
	}

	cfStr := C.CFStringRef(idPtr)

	var buf [256]C.char
	ok := C.CFStringGetCString(
		cfStr,
		(*C.char)(unsafe.Pointer(&buf[0])),
		256,
		C.kCFStringEncodingUTF8,
	)
	if ok == 0 {
		return InputMethodUnknown, ""
	}

	id := C.GoString(&buf[0])

	switch {
	case strings.HasPrefix(id, ENGLISH_INPUT_METHOD_ID):
		return InputMethodEnglish, id
	case strings.HasPrefix(id, CHINESE_INPUT_METHOD_ID):
		return InputMethodIME, id
	default:
		return InputMethodUnknown, id
	}
}

// SwitchInputMethod switches macOS input method by InputSourceID
func SwitchInputMethod(id string) error {
	cfStr := C.CFStringCreateWithCString(
		C.kCFAllocatorDefault,
		C.CString(id),
		C.kCFStringEncodingUTF8,
	)
	if cfStr == 0 {
		return errors.New("failed to create CFString")
	}
	defer C.CFRelease(C.CFTypeRef(cfStr))

	// Create dictionary: { kTISPropertyInputSourceID: id }
	keys := []C.CFStringRef{C.kTISPropertyInputSourceID}
	values := []C.CFTypeRef{C.CFTypeRef(cfStr)}

	dict := C.CFDictionaryCreate(
		C.kCFAllocatorDefault,
		(*unsafe.Pointer)(unsafe.Pointer(&keys[0])),
		(*unsafe.Pointer)(unsafe.Pointer(&values[0])),
		1,
		&C.kCFTypeDictionaryKeyCallBacks,
		&C.kCFTypeDictionaryValueCallBacks,
	)
	if dict == 0 {
		return errors.New("failed to create dictionary")
	}
	defer C.CFRelease(C.CFTypeRef(dict))

	list := C.TISCreateInputSourceList(dict, C.Boolean(0))
	if list == 0 {
		return errors.New("input source not found")
	}
	defer C.CFRelease(C.CFTypeRef(list))

	if C.CFArrayGetCount(list) == 0 {
		return errors.New("no matching input source")
	}

	source := C.TISInputSourceRef(
		C.CFArrayGetValueAtIndex(list, 0),
	)

	if C.TISSelectInputSource(source) != 0 {
		return errors.New("failed to select input source")
	}

	return nil
}

// InputMethodID returns the id of a IMEType.
func InputMethodID(t InputMethodType) (string, error) {
	switch t {
	case InputMethodEnglish:
		return ENGLISH_INPUT_METHOD_ID, nil
	case InputMethodIME:
		return CHINESE_INPUT_METHOD_ID, nil
	default:
		return "", errors.New("unknown input method type")
	}
}
