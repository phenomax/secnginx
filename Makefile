.PHONY =  clean, lint

GO_FILES := $(shell go list -f '{{range $$index, $$element := .GoFiles}}{{$$.Dir}}/{{$$element}}{{"\n"}}{{end}}' ./... | grep -v '/vendor/')

default: clean, lint

clean:
	rm -rf dist/ builds/

lint:
	golint $(GO_FILES)
