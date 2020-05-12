build:	test
	go build -o $(GOPATH)/bin/git-appraise-web git-appraise-web/git-appraise-web.go

test:	vet
	go test ./...

vet:	fmt
	go vet ./...

fmt:	assets
	gofmt -w ./

assets: deps FORCE
	rm assets/*~ 2>/dev/null || true
	$(GOPATH)/bin/go-bindata -pkg assets -o third_party/assets/assets.go assets/

deps:
	go get -u github.com/jteeuwen/go-bindata/...

FORCE:
