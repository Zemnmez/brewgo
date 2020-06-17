brewgo
======

Install go modules as brew formulae.

The intent of this tool is to have go modules installed to brew as brew formulae,
so that brew bundle dump can pick them up. So far, it can just generate formulae
for any go package. You can do this now via:

```bash
go run github.com/zemnmez/brewgo -print github.com/zemnmez/aquatone
```

