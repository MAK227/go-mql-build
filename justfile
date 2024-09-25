freeze-source:
    rm assets/freeze.png
    freeze --execute "go-mql-build -h"
    mv freeze.png assets/freeze.png

release MSG VERSION: demo
    @just commit "{{MSG}}"
    git tag "v{{VERSION}}"
    git push origin "v{{VERSION}}"
    goreleaser release --clean
    go install github.com/MAK227/go-mql-build@v{{VERSION}}
    @just commit "Updating freeze.png for v{{VERSION}}"

test: build
    go-mql-build-local -h

build:
    go mod tidy
    go build -o go-mql-build-local .
    mv go-mql-build-local ~/.local/bin

commit MSG:
    @just freeze-source
    git add .
    git commit -m "{{MSG}}"
    git push

demo:
    vhs assets/go-mql-successful.tape
    vhs assets/go-mql-fail.tape
