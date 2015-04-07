FROM golang:1.3

COPY . /go/src/github.com/lookify/town
WORKDIR /go/src/github.com/lookify/town

ENV GOPATH /go/src/github.com/lookify/town/Godeps/_workspace:$GOPATH
RUN CGO_ENABLED=0 go install -v -a -tags netgo -ldflags "-w -X github.com/lookify/town/version.GITCOMMIT `git rev-parse --short HEAD`"

EXPOSE 2375

VOLUME $HOME/.town

ENTRYPOINT ["town"]
CMD ["--help"]