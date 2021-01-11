package erk

// ExportedErkable is an exported readonly version of the Erkable interface.
type ExportedErkable interface {
	ErrorMessage() string
	ErrorKind() string
	ErrorParams() Params
}

// Exportable errors that support being exported to a JSON marshal friendly format.
type Exportable interface {
	ExportRawMessage() string
	Export() ExportedErkable
}

// BaseExport error that satisfies the ExportedErkable interface and is useful for JSON marshalling.
type BaseExport struct {
	Kind    string `json:"kind"`
	Message string `json:"message"`
	Params  Params `json:"params,omitempty"`
}

// ErrorMessage returns the error message.
func (e *BaseExport) ErrorMessage() string {
	return e.Message
}

// ErrorKind returns the error kind.
func (e *BaseExport) ErrorKind() string {
	return e.Kind
}

// ErrorParams returns the error params.
func (e *BaseExport) ErrorParams() Params {
	return e.Params
}

// Export creates a visible copy of the error that can be used outside the erk package.
// A common use case is marshalling the error to JSON.
// If err is not an erk.Erkable, it is wrapped first.
func Export(err error) ExportedErkable {
	return ToErk(err).Export()
}
