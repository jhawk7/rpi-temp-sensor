FROM golang:1.17-alpine AS build
WORKDIR /build
COPY go.mod ./
COPY go.sum ./
RUN go mod download
COPY *.go ./
RUN go build -o thermo

FROM golang:1.17-alpine
WORKDIR /app
COPY --from=build thermo thermo
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["thermo"]
