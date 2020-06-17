package brewgo

import (
	"encoding"
	"errors"
	"fmt"
	"path"
)

func MarshalTextAsString(f encoding.TextMarshaler) (text string, err error) {
	bt, err := f.MarshalText()
	text = string(bt)
	return
}

// $GOPATH/<Module>/@v
type proxyPathSpecial struct {
	pkgDescriptor
}

func (p proxyPathSpecial) MarshalText() (text []byte, err error) {
	return []byte(fmt.Sprintf(
		"/%s/@%s",
		path.Clean(string(p.Module)),
		path.Clean(string(p.special)),
	)), nil
}

type proxyPathGoMod struct {
	pkgDescriptor
}

func (p proxyPathGoMod) MarshalText() (text []byte, err error) {
	if len(p.version) == 0 {
		err = errors.New("must have a specified version")
		return
	}

	return []byte(fmt.Sprintf(
		"%s/@v/%s.mod",
		path.Clean(string(p.Module)),
		path.Clean(string(p.version)),
	)), nil
}

type proxyPathZip struct {
	pkgDescriptor
}

func (p proxyPathZip) MarshalText() (text []byte, err error) {
	if len(p.version) == 0 {
		err = errors.New("must have a specified version")
		return
	}

	return []byte(fmt.Sprintf(
		"/%s/@v/%s.zip",
		path.Clean(string(p.Module)),
		path.Clean(string(p.version)),
	)), nil
}
