#FROM golang:1.17-alpine AS build
#RUN apk add build-base
#WORKDIR /build
#COPY . ./
#COPY go.mod ./
#COPY go.sum ./
#RUN go mod download
#RUN go build -o thermo

FROM golang:1.17-alpine
WORKDIR /app
#COPY --from=build thermo thermo
COPY thermo ./
EXPOSE 8080
USER nonroot:nonroot
ENTRYPOINT ["/thermo"]
