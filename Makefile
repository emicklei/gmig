clean:
	rm -rf target

build:
	docker run --rm -it -v "${PWD}":/go/src/github.com/emicklei/gmig -w /go/src/github.com/emicklei/gmig golang:1.10 make build_inside

build_inside:
	cd /go/src/github.com/emicklei/gmig && \
	rm -rf target && \
	mkdir -p target/windows && \
	mkdir -p target/darwin && \
	mkdir -p target/linux && \
	GOOS=windows go build -o target/windows/gmig.exe && \
	GOOS=darwin go build -o target/darwin/gmig && \
	GOOS=linux go build -o target/linux/gmig && \
	chmod +x -R target

zip:
	zip target/darwin/gmig.zip target/darwin/gmig && \
	zip target/linux/gmig.zip target/linux/gmig && \
	zip target/windows/gmig.zip target/windows/gmig.exe

# go get github.com/aktau/github-release
# export GITHUB_TOKEN=...
.PHONY: createrelease
createrelease:
	github-release info -u emicklei -r gmig
	github-release release \
		--user emicklei \
		--repo gmig \
		--tag $(shell git tag -l --points-at HEAD) \
		--name "gmig" \
		--description "gmig - google infrastructure-as-code tool"

.PHONY: uploadrelease
uploadrelease:
	github-release upload \
		--user emicklei \
		--repo gmig \
		--tag $(shell git tag -l --points-at HEAD) \
		--name "gmig-Linux-x86_64.zip" \
		--file target/linux/gmig.zip

	github-release upload \
		--user emicklei \
		--repo gmig \
		--tag $(shell git tag -l --points-at HEAD) \
		--name "gmig-Darwin-x86_64.zip" \
		--file target/darwin/gmig.zip

	github-release upload \
		--user emicklei \
		--repo gmig \
		--tag $(shell git tag -l --points-at HEAD) \
		--name "gmig-Windows-x86_64.zip" \
		--file target/windows/gmig.zip