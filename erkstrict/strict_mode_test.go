package erkstrict_test

import (
	"os"
	"testing"

	"github.com/JosiahWitt/ensure"
	"github.com/JosiahWitt/ensure/ensurepkg"
	"github.com/JosiahWitt/erk/erkstrict"
)

func TestIsStrictMode(t *testing.T) {
	ensure := ensure.New(t)

	expectTrue := func(ensure ensurepkg.Ensure) func() {
		return func() {
			erkstrict.UnsetStrictMode()

			ensure(erkstrict.IsStrictMode()).IsTrue()
		}
	}

	expectFalse := func(ensure ensurepkg.Ensure) func() {
		return func() {
			erkstrict.UnsetStrictMode()

			ensure(erkstrict.IsStrictMode()).IsFalse()
		}
	}

	ensure.Run("not in tests", func(ensure ensurepkg.Ensure) {
		originalArgs := make([]string, len(os.Args))
		copy(originalArgs, os.Args)
		os.Args = []string{"testing"}
		defer func() {
			os.Args = originalArgs
		}()

		ensure.Run("with unset erk strict mode", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("", expectFalse(ensure))
		})

		ensure.Run("with erk strict disabled", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("false", expectFalse(ensure))
		})

		ensure.Run("with erk strict enabled", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("true", expectTrue(ensure))
		})
	})

	ensure.Run("in tests", func(ensure ensurepkg.Ensure) {
		originalArgs := make([]string, len(os.Args))
		copy(originalArgs, os.Args)
		os.Args = []string{"testing", "-test.thing=true"}
		defer func() {
			os.Args = originalArgs
		}()

		ensure.Run("with unset erk strict mode", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("", expectTrue(ensure))
		})

		ensure.Run("with erk strict disabled", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("false", expectFalse(ensure))
		})

		ensure.Run("with erk strict enabled", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("true", expectTrue(ensure))
		})
	})

	ensure.Run("when not changing args (in test)", func(ensure ensurepkg.Ensure) {
		ensure.Run("with unset erk strict mode", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("", expectTrue(ensure))
		})

		ensure.Run("with erk strict disabled", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("false", expectFalse(ensure))
		})

		ensure.Run("with erk strict enabled", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("true", expectTrue(ensure))
		})
	})

	ensure.Run("strict mode is cached", func(ensure ensurepkg.Ensure) {
		ensure.Run("with erk strict disabled", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("false", func() {
				erkstrict.UnsetStrictMode()

				ensure(erkstrict.IsStrictMode()).IsFalse()

				withErkStrictEnv("true", func() {
					ensure(erkstrict.IsStrictMode()).IsFalse() // Does not change
				})
			})
		})

		ensure.Run("with erk strict enabled", func(ensure ensurepkg.Ensure) {
			withErkStrictEnv("true", func() {
				erkstrict.UnsetStrictMode()

				ensure(erkstrict.IsStrictMode()).IsTrue()

				withErkStrictEnv("false", func() {
					ensure(erkstrict.IsStrictMode()).IsTrue() // Does not change
				})
			})
		})
	})
}

func TestUnsetStrictMode(t *testing.T) {
	ensure := ensure.New(t)

	withErkStrictEnv("false", func() {
		erkstrict.UnsetStrictMode()

		ensure(erkstrict.IsStrictMode()).IsFalse()

		withErkStrictEnv("true", func() {
			ensure(erkstrict.IsStrictMode()).IsFalse() // Does not change

			erkstrict.UnsetStrictMode()
			ensure(erkstrict.IsStrictMode()).IsTrue() // Changes after strict mode unset
		})
	})
}

func TestSetStrictMode(t *testing.T) {
	ensure := ensure.New(t)

	ensure.Run("set from true to false", func(ensure ensurepkg.Ensure) {
		withErkStrictEnv("true", func() {
			erkstrict.SetStrictMode(false)

			ensure(erkstrict.IsStrictMode()).IsFalse()
		})
	})

	ensure.Run("set from false to true", func(ensure ensurepkg.Ensure) {
		withErkStrictEnv("false", func() {
			erkstrict.SetStrictMode(true)

			ensure(erkstrict.IsStrictMode()).IsTrue()
		})
	})
}

func withErkStrictEnv(value string, fn func()) {
	strict, isSet := os.LookupEnv("ERK_STRICT_MODE")
	if value != "" {
		os.Setenv("ERK_STRICT_MODE", value)
	} else {
		os.Unsetenv("ERK_STRICT_MODE")
	}

	fn()

	if isSet {
		os.Setenv("ERK_STRICT_MODE", strict)
	} else {
		os.Unsetenv("ERK_STRICT_MODE")
	}
}
