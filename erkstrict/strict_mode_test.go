package erkstrict_test

import (
	"os"
	"testing"

	"github.com/JosiahWitt/erk/erkstrict"
	"github.com/matryer/is"
)

func TestIsStrictMode(t *testing.T) {
	expectTrue := func(t *testing.T) func() {
		return func() {
			is := is.New(t)
			erkstrict.UnsetStrictMode()

			is.True(erkstrict.IsStrictMode())
		}
	}

	expectFalse := func(t *testing.T) func() {
		return func() {
			is := is.New(t)
			erkstrict.UnsetStrictMode()

			is.Equal(erkstrict.IsStrictMode(), false)
		}
	}

	t.Run("not in tests", func(t *testing.T) {
		originalArgs := make([]string, len(os.Args))
		copy(originalArgs, os.Args)
		os.Args = []string{"testing"}
		defer func() {
			os.Args = originalArgs
		}()

		t.Run("with unset erk strict mode", func(t *testing.T) {
			withErkStrictEnv("", expectFalse(t))
		})

		t.Run("with erk strict disabled", func(t *testing.T) {
			withErkStrictEnv("false", expectFalse(t))
		})

		t.Run("with erk strict enabled", func(t *testing.T) {
			withErkStrictEnv("true", expectTrue(t))
		})
	})

	t.Run("in tests", func(t *testing.T) {
		originalArgs := make([]string, len(os.Args))
		copy(originalArgs, os.Args)
		os.Args = []string{"testing", "-test.thing=true"}
		defer func() {
			os.Args = originalArgs
		}()

		t.Run("with unset erk strict mode", func(t *testing.T) {
			withErkStrictEnv("", expectTrue(t))
		})

		t.Run("with erk strict disabled", func(t *testing.T) {
			withErkStrictEnv("false", expectFalse(t))
		})

		t.Run("with erk strict enabled", func(t *testing.T) {
			withErkStrictEnv("true", expectTrue(t))
		})
	})

	t.Run("when not changing args (in test)", func(t *testing.T) {
		t.Run("with unset erk strict mode", func(t *testing.T) {
			withErkStrictEnv("", expectTrue(t))
		})

		t.Run("with erk strict disabled", func(t *testing.T) {
			withErkStrictEnv("false", expectFalse(t))
		})

		t.Run("with erk strict enabled", func(t *testing.T) {
			withErkStrictEnv("true", expectTrue(t))
		})
	})

	t.Run("strict mode is cached", func(t *testing.T) {
		t.Run("with erk strict disabled", func(t *testing.T) {
			withErkStrictEnv("false", func() {
				is := is.New(t)
				erkstrict.UnsetStrictMode()

				is.Equal(erkstrict.IsStrictMode(), false)

				withErkStrictEnv("true", func() {
					is.Equal(erkstrict.IsStrictMode(), false) // Does not change
				})
			})
		})

		t.Run("with erk strict enabled", func(t *testing.T) {
			withErkStrictEnv("true", func() {
				is := is.New(t)
				erkstrict.UnsetStrictMode()

				is.True(erkstrict.IsStrictMode())

				withErkStrictEnv("false", func() {
					is.True(erkstrict.IsStrictMode()) // Does not change
				})
			})
		})
	})
}

func TestUnsetStrictMode(t *testing.T) {
	withErkStrictEnv("false", func() {
		is := is.New(t)
		erkstrict.UnsetStrictMode()

		is.Equal(erkstrict.IsStrictMode(), false)

		withErkStrictEnv("true", func() {
			is.Equal(erkstrict.IsStrictMode(), false) // Does not change

			erkstrict.UnsetStrictMode()
			is.True(erkstrict.IsStrictMode()) // Changes after strict mode unset
		})
	})
}

func TestSetStrictMode(t *testing.T) {
	t.Run("set from true to false", func(t *testing.T) {
		withErkStrictEnv("true", func() {
			is := is.New(t)
			erkstrict.SetStrictMode(false)

			is.Equal(erkstrict.IsStrictMode(), false)
		})
	})

	t.Run("set from false to true", func(t *testing.T) {
		withErkStrictEnv("false", func() {
			is := is.New(t)
			erkstrict.SetStrictMode(true)

			is.True(erkstrict.IsStrictMode())
		})
	})
}

func withErkStrictEnv(value string, fn func()) {
	strict, isSet := os.LookupEnv("ERK_STRICT")
	if value != "" {
		os.Setenv("ERK_STRICT", value)
	} else {
		os.Unsetenv("ERK_STRICT")
	}

	fn()

	if isSet {
		os.Setenv("ERK_STRICT", strict)
	} else {
		os.Unsetenv("ERK_STRICT")
	}
}
