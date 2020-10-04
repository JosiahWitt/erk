package erkmock_test

import (
	"errors"
	"testing"

	"github.com/JosiahWitt/erk"
	"github.com/JosiahWitt/erk/erkmock"
	"github.com/matryer/is"
)

func TestFrom(t *testing.T) {
	t.Run("with erk error", func(t *testing.T) {
		is := is.New(t)

		erkErr := erk.New(TestKind{}, "my message")
		m := erkmock.From(erkErr)
		is.Equal(m.(erk.Kindable).Kind(), TestKind{})
	})

	t.Run("with non-erk error", func(t *testing.T) {
		is := is.New(t)

		defer func() {
			if res := recover(); res != nil {
				str, ok := res.(string)
				is.True(ok)
				is.Equal(str, "erkmock.From only supports mocking erk.Erkable errors")
			}
		}()

		err := errors.New("going to explode")
		erkmock.From(err) // nolint:errcheck // We shouldn't be able to test the output of this function
		is.Fail()         // Expected panic
	})
}
