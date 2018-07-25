INGESTOR=bin/ingestor

ifeq ($(TRAVIS), true)
	CGO_ENABLED := 0
else
	CGO_ENABLED := 1
endif

build:
	CGO_ENABLED=${CGO_ENABLED} go build -o ${INGESTOR} *.go

run:
	bin/ingestor
