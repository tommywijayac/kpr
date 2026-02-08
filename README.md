Access it at `https://tommywijayac.github.io/kpr/`

Simple KPR calculator in Indonesia, addressing the lack of available ones for tiered interest.

## Build & Run locally

### Install gopherjs
```
go install golang.org/dl/go1.19.13@latest
go1.19.13 download
go1.19.13 install github.com/gopherjs/gopherjs@v1.19.0-beta2
export GOPHERJS_GOROOT="$(go1.19.13 env GOROOT)"
```

### Build and run
```
gopherjs build . && gopherjs serve
```