package brewgo

import (
	"bytes"
	"errors"
	"fmt"
	"os/exec"
)

// ErrEmptyEnv is returned when the go env requested is empty
type ErrEmptyEnv struct {
	name string
}

func (e ErrEmptyEnv) Error() string {
	return fmt.Sprintf("env %s is empty", e.name)
}

// Env returns the go environment variable specified.
// Ideally this'd work by importing part of the Go toolchain
// but for legacy reasons, doing so is unstable
func Env(env string) (value []byte, err error) {
	value, err = exec.Command("go", "env", env).Output()
	if err != nil {
		return
	}

	value = bytes.TrimSpace(value)

	if len(value) == 0 {
		err = ErrEmptyEnv{env}
	}

	return
}

func EnvGoproxy() (value []byte, err error) {
	v, err := Env("GOPROXY")
	if err != nil {
		return
	}

	splits := bytes.Split(v, []byte{','})
	if len(splits) < 1 {
		err = errors.New("GOPROXY not set")
		return
	}

	proxy := splits[0]

	if bytes.Equal(proxy, []byte("direct")) {
		err = errors.New("GOPROXY must be set to something other than 'direct'")
		return
	}

	return proxy, err
}
