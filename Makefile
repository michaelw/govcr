DEPS:=$$(go list -f '{{range .TestImports}}{{.}} {{end}}' ./...)
MAINFILES=$$(go list -f '{{join .GoFiles " "}}')

deps:
	go get -d -t -v ./... $(DEPS)

test: deps
	go test -cover -race -parallel 100 ./...

examples: deps
	go run ./examples/*.go

lint: deps
	@-echo "NOTE: some linters (gotype) require a recent 'go install'"
	go get -u github.com/alecthomas/gometalinter
	gometalinter --install --update
	gometalinter --deadline=300s --disable=dupl --concurrency=4

goconvey:
	@-killall goconvey
	@-echo "NOTE: you may be required to perform a recent 'go install'"
	${GOPATH}/bin/goconvey -depth=100 -timeout=600s -excludedDirs=.git,.vscode,.idea -packages=2 -cover -poll=5000ms -port=6020 1>/dev/null 2>&1 &

