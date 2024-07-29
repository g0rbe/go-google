# go-google

[![Go Report Card](https://goreportcard.com/badge/github.com/g0rbe/go-google)](https://goreportcard.com/report/github.com/g0rbe/go-google)
[![Go Reference](https://pkg.go.dev/badge/github.com/g0rbe/go-google.svg)](https://pkg.go.dev/github.com/g0rbe/go-google)

Golang module to Google APIs.

- `Pagespeed Insight` / `Lighthouse`: [API Reference](https://developers.google.com/speed/docs/insights/rest/v5/pagespeedapi/runpagespeed)

Get:
```bash
go get github.com/g0rbe/go-google@latest
```

Get the latest tag (if Go module proxy is not updated):
```bash
go get "github.com/g0rbe/go-google@$(curl -s 'https://api.github.com/repos/g0rbe/go-google/tags' | jq -r '.[0].name')"
```

Get the latest commit (if Go module proxy is not updated):
```bash
go get "github.com/g0rbe/go-google@$(curl -s 'https://api.github.com/repos/g0rbe/go-google/commits' | jq -r '.[0].sha')"
```

## TODO

- `Common errors`