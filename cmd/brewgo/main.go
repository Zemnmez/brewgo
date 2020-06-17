package main // import "github.com/zemnmez/brewgo/cmd/brewgo"

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"

	"github.com/zemnmez/brewgo"
)

func main() {
	if err := do(context.Background()); err != nil {
		panic(err)
	}
}

type Program struct {
	Args    []string
	Stdout  io.Writer
	Stderr  io.Writer
	Targets []string
	Print   bool
}

func (p *Program) SetFlags(ctx context.Context, set *flag.FlagSet) (err error) {
	set.BoolVar(&p.Print, "print", false, "Print to stdout instead of installing the package via brew")
	return
}

func (p *Program) Parse(ctx context.Context, args []string, set *flag.FlagSet) (err error) {
	if err = set.Parse(args[1:]); err != nil {
		return
	}

	p.Targets = set.Args()
	return
}

func (p *Program) Run(ctx context.Context) (err error) {
	flagSet := flag.NewFlagSet("", flag.ContinueOnError)
	flagSet.SetOutput(p.Stderr)
	flagSet.Usage = func() {
		fmt.Fprintf(flagSet.Output(), "Usage of %s:\n", p.Args[0])
		fmt.Fprintf(
			flagSet.Output(),
			"%s [packages...] [args]\n",
			p.Args[0],
		)
		flagSet.PrintDefaults()
	}

	if err = p.SetFlags(ctx, flagSet); err != nil {
		return
	}

	if err = p.Parse(ctx, p.Args, flagSet); err != nil {
		return
	}

	if p.Print {
		return p.PrintCommand(ctx)
	}

	return p.InstallCommand(ctx)
}

func (p *Program) PrintCommand(ctx context.Context) (err error) {
	if len(p.Targets) > 1 {
		return errors.New("cannot print more than 1 target")
	}

	if len(p.Targets) == 0 {
		return errors.New("must have at least 1 target to print")
	}

	inf, err := brewgo.GetInfo([]byte(p.Targets[0]))
	if err != nil {
		return
	}

	if _, err = io.Copy(p.Stdout, inf); err != nil {
		return
	}

	return
}

func (p *Program) InstallCommand(ctx context.Context) (err error) {
	return
}

func do(ctx context.Context) (err error) {
	return (&Program{
		Stdout: os.Stdout,
		Args:   os.Args,
		Stderr: os.Stderr,
	}).Run(ctx)
}
