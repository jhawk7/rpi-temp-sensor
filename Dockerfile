FROM golang:1.17-alpine AS build
RUN apk add build-base
WORKDIR /build
COPY . ./
RUN go mod download
RUN go build -o /bin/thermo

FROM golang:1.17-alpine
WORKDIR /app
COPY --from=build /bin/thermo /bin/thermo
EXPOSE 8080
CMD ["/bin/thermo"]
