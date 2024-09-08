freeze-source:
    freeze --execute "go-mql-build -h" 

release MSG VERSION:
    just commit "{{MSG}}"
    git tag "v{{VERSION}}"
    git push origin "v{{VERSION}}"
    goreleaser release --clean

test:
    go mod tidy
    go build .
    mv go-mql-build ~/.local/bin
    go-mql-build -h

build:
    go mod tidy
    go build .
    mv go-mql-build ~/.local/bin

commit MSG:
    just freeze-source
    git add .
    git commit -m "{{MSG}}"
    git push
