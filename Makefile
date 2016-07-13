build:	test
	go build -o $(GOPATH)/bin/git-appraise-web git-appraise-web/git-appraise-web.go

test:	vet
	go test ./...

vet:	fmt
	go vet ./...

fmt:	assets
	gofmt -w ./

assets: FORCE
	go get -u github.com/jteeuwen/go-bindata/...
	rm assets/*~ 2>/dev/null || true
	$(GOPATH)/bin/go-bindata -modtime 1 -pkg assets -o third_party/assets/assets.go assets/

FORCE:
