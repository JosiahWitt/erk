package erkmock_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erkmock"
)

func TestFrom(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("with erk error", func(ensure ensurepkg.Ensure) {
		erkErr := erk.New(TestKind{}, "my message")
		m := erkmock.From(erkErr)
		ensure(m.(erk.Kindable).Kind()).Equals(TestKind{})
		ensure(m.(erk.Exportable).ExportRawMessage()).Equals("my message")
	})

	ensure.Run("with non-erk error", func(ensure ensurepkg.Ensure) {
		defer func() {
			if res := recover(); res != nil {
				str, ok := res.(string)
				ensure(ok).IsTrue()
				ensure(str).Equals("erkmock.From only supports mocking erk.Erkable errors")
			}
		}()

		err := errors.New("going to explode")
		erkmock.From(err) // nolint:errcheck // We shouldn't be able to test the output of this function
		ensure.Failf("Expected panic, so this line should not be reached")
	})
}
