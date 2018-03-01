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
	chmod +x -R target &&
