FROM golang:1.24 AS build
WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
RUN go mod tidy
RUN go mod verify

COPY . ./

ENV CGO_ENABLED=0
RUN go build -trimpath -o /bin/app ./cmd/api

FROM gcr.io/distroless/static-debian12:nonroot
COPY --from=build /bin/app /bin/app
ENTRYPOINT [ "/bin/app" ]
