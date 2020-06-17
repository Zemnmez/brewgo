package brewgo

import (
	"bytes"
	"crypto/sha256"
	"encoding"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"text/template"
)

var tmpl = template.Must(template.New("formula").Funcs(template.FuncMap{
	"hex": func(b []byte) string {
		return hex.EncodeToString(b)
	},
	"title": func(b []byte) []byte {
		return bytes.Title(b)
	},
}).Parse(`
class {{title .URI.Name | printf "%s"}} < Formula
	depends_on "golang"
	homepage {{printf "%+q" .Homepage}}
	version {{printf "%+q" .Version}}
	url {{printf "%+q" .URL}}
	sha256 {{hex .Sha256 | printf "%+q"}}

	def install
		# brew tries to be helpful here and
		# cds into the first directory, which actually
		# messes us up
		cd "../" + {{printf "%+q" .URI}} do
			ENV["GOBIN"] = bin
			system "go", "get", "./..."
		end
	end
end

`))

// Info is all the data needed
// to make a Brew package
type Info struct {
	URI      pkgDescriptor
	Homepage *url.URL
	URL      *url.URL
	Sha256   []byte
	Version  []byte
}

func (Info) Read([]byte) (n int, err error) {
	return 0, errors.New("this is an io.WriterTo, not an io.Writer")
}

type writeCounter struct {
	io.Writer
	count int64
}

func (w *writeCounter) Write(b []byte) (n int, err error) {
	n, err = w.Writer.Write(b)
	w.count += int64(n)
	return
}

// WriteTo writes the brew package to a writer
func (i Info) WriteTo(w io.Writer) (n int64, err error) {
	var counter = writeCounter{Writer: w}
	err = tmpl.Execute(&counter, i)
	n = counter.count

	return
}

type pkgDescriptor struct {
	Module  []byte
	version []byte
	// @latest, @upgrade etc
	special []byte
}

func (p pkgDescriptor) Name() (name []byte, err error) {
	return []byte(path.Base(string(p.Module))), nil
}

func (p *pkgDescriptor) UnmarshalText(text []byte) (err error) {
	splits := bytes.Split(text, []byte{'@'})
	p.Module = splits[0]
	if len(splits) > 1 {
		switch {
		case len(splits[1]) > 0 && splits[1][0] == 'v':
			p.version = splits[1]
		default:
			p.special = splits[1]
		}
	}

	return
}

var _ encoding.TextMarshaler = pkgDescriptor{}

func (p pkgDescriptor) String() string {
	text, err := p.MarshalText()
	if err != nil {
		panic(err)
	}

	return string(text)
}

func (p pkgDescriptor) MarshalText() (text []byte, err error) {
	// should support special here, but i never use it so...
	text = bytes.Join([][]byte{p.Module, p.version}, []byte{'@'})
	return
}

// GetInfo returns info for a go package, that can be used to
// make a brew formula
func GetInfo(pkg []byte) (inf Info, err error) {
	var pkgURI pkgDescriptor
	if err = pkgURI.UnmarshalText(pkg); err != nil {
		return
	}

	// if no version is specified and no special (@latest)
	// is specified, default to latest
	if len(pkgURI.version) == 0 || len(pkgURI.special) == 0 {

		pkgURI.special = []byte("latest")
	}

	goproxyBt, err := EnvGoproxy()
	if err != nil {
		return
	}

	goproxyURL, err := url.Parse(string(goproxyBt))
	if err != nil {
		return
	}

	//gosumdb, err := Env("GOSUMDB")
	//if err != nil {
	//	return
	//}

	//gosumURL, err := url.Parse(string(gosumdb))
	// -- we just calculate this ourselves

	//if err != nil {
	//	return
	//}

	// get latest version info
	var versionInfo struct {
		Version string `json:"version"`
	}

	// copy
	var versionInfoURL = *goproxyURL

	versionInfoURL.Path, err = MarshalTextAsString(proxyPathSpecial{pkgURI})
	if err != nil {
		return
	}

	rsp, err := http.Get(versionInfoURL.String())
	if err != nil {
		return
	}

	if err = json.NewDecoder(rsp.Body).Decode(&versionInfo); err != nil {
		return
	}

	latestVersion := []byte(versionInfo.Version)
	if len(latestVersion) == 0 || latestVersion == nil {
		err = errors.New("missing latest version")
		return
	}

	pkgURI.version = []byte(latestVersion)

	// calculate sha256 checksum of archive
	var archiveURL = *goproxyURL

	archiveURL.Path, err = MarshalTextAsString(proxyPathZip{pkgURI})
	if err != nil {
		return
	}

	rsp, err = http.Get(archiveURL.String())
	if err != nil {
		return
	}

	var hash = sha256.New()
	if _, err = io.Copy(hash, rsp.Body); err != nil {
		return
	}

	return Info{
		URI: pkgURI,
		Homepage: &url.URL{
			Scheme: "https",
			Host:   "godoc.org",
			Path:   string(pkgURI.Module),
		},

		URL: &archiveURL,

		Sha256: hash.Sum(nil),

		Version: pkgURI.version,
	}, nil

}
