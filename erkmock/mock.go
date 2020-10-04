package erkmock

import "github.com/JosiahWitt/erk"

// Mock erk.Erkable implementation.
// Mock does contain an error message.
// Setting parameters on a mock modify the mock in place.
// Thus, it is recommended to create a new mock instead of using the same one multiple times.
type Mock struct {
	kind   erk.Kind
	params erk.Params
}

var _ erk.Erkable = &Mock{}

// For a given erk kind, create a mock error.
func For(kind erk.Kind) error {
	return &Mock{
		kind:   kind,
		params: erk.Params{},
	}
}

// Error returns "MOCK: <kind string>".
func (m *Mock) Error() string {
	return "MOCK: " + erk.GetKindString(m)
}

// Export the mock error.
func (m *Mock) Export() erk.ExportedErkable {
	return &erk.ExportedError{
		BaseExport: erk.BaseExport{
			Kind:    erk.GetKindString(m),
			Message: m.Error(),
			Params:  m.params,
		},
	}
}

// Is implements the Go 1.13+ Is interface for use with errors.Is.
func (m *Mock) Is(err error) bool {
	return erk.IsKind(err, m.kind)
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
