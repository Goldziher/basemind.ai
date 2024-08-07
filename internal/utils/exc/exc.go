package exc

import (
	"errors"
	"github.com/rs/zerolog/log"
	"runtime"
	"strings"
)

// MustResult - panics if the error is not nil, otherwise returns the result.
func MustResult[T any](value T, err error, messages ...string) T {
	if err != nil {
		pc, _, lineNumber, _ := runtime.Caller(1)
		details := runtime.FuncForPC(pc)

		log.Error().
			Err(err).
			Str("caller", details.Name()).
			Int("line number", lineNumber).
			Msg(strings.Join(messages, " "))
		panic(err)
	}
	return value
}

// Must - panics if the error is not nil.
func Must(err error, messages ...string) {
	if err != nil {
		log.Error().Err(err).Msg(strings.Join(messages, " "))
		panic(err)
	}
}

// LogIfErr - logs the error if it is not nil.
func LogIfErr(err error, messages ...string) {
	if err != nil {
		log.Error().Err(err).Msg(strings.Join(messages, " "))
	}
}

// ReturnNotNil - panics if the value is nil, otherwise returns the value.
func ReturnNotNil[T any](value *T, messages ...string) *T {
	if value == nil {
		panic(errors.New(strings.Join(messages, " ")))
	}
	return value
}

// ReturnAnyErr - returns the first error in the list of values, if any.
func ReturnAnyErr(values ...any) error {
	for _, value := range values {
		if err, ok := value.(error); ok && err != nil {
			return err
		}
	}
	return nil
}
