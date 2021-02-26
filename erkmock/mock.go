package erkmock

import (
	"fmt"

	"github.com/JosiahWitt/erk"
)

// Mock erk.Erkable implementation.
// Setting parameters on a mock modify the mock in place.
// Thus, it is recommended to create a new mock instead of using the same one multiple times.
type Mock struct {
	kind    erk.Kind
	params  erk.Params
	message string
}

var _ erk.Erkable = &Mock{}

// For a given erk kind, create a mock error.
func For(kind erk.Kind) error {
	return &Mock{
		kind:   kind,
		params: erk.Params{},
	}
}

// SetMessage on the mock.
func (m *Mock) SetMessage(message string) {
	m.message = message
}

// Error returns the error kind, message, and parameters formatted as a string.
func (m *Mock) Error() string {
	if m.message == "" {
		return fmt.Sprintf("{KIND: \"%s\", PARAMS: %+v}", erk.GetKindString(m), m.Params())
	}

	return fmt.Sprintf("{KIND: \"%s\", RAW MESSAGE: \"%s\", PARAMS: %+v}", erk.GetKindString(m), m.ExportRawMessage(), m.Params())
}

// ExportRawMessage without executing the template.
func (m *Mock) ExportRawMessage() string {
	return m.message
}

// Export the mock error.
func (m *Mock) Export() erk.ExportedErkable {
	return &erk.BaseExport{
		Kind:    erk.GetKindString(m),
		Message: m.Error(),
		Params:  m.params,
	}
}

// Is implements the Go 1.13+ Is interface for use with errors.Is.
//
// If the mock has no message set, only the error kinds are compared.
// Otherwise, the error kinds and messages are compared.
func (m *Mock) Is(err error) bool {
	isKind := erk.IsKind(err, m.kind)
	if !isKind || m.message == "" {
		return isKind
	}

	erkable, isErkable := err.(erk.Erkable)
	return isErkable && erkable.ExportRawMessage() == m.message
}

// Kind returns the mock error kind.
func (m *Mock) Kind() erk.Kind {
	return m.kind
}

// WithParams modifies the mock error params in place (instead of on a clone).
// Note that this behavior is different, since it allows checking
// on the original mock if any errors were set.
func (m *Mock) WithParams(params erk.Params) error {
	for k, v := range params {
		m.params[k] = v
	}

	return m
}

// Params set on the mock error.
func (m *Mock) Params() erk.Params {
	return m.params
}
