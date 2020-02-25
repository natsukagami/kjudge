# kjudge

## Build Instructions

External Dependencies:

```yaml
cc:      anything that compiles SQLite 3
go:      >= 1.13
node.js: >= 9
yarn:    >= 1
```

Build steps:

```sh
go generate
go build -o kjudge cmd/kjudge/main.go
```
