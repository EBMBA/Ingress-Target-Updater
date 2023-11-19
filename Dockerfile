FROM --platform=linux/amd64 golang:1.21.1 as build

WORKDIR /go/src/app

COPY go.mod go.sum ./

RUN go mod download

COPY *.go ./

RUN CGO_ENABLED=0 GOOS=linux go build -o /go/src/app

FROM gcr.io/distroless/static-debian11
COPY --from=build /go/bin/app /
CMD [ "/app" ]
