FROM golang:1.17 AS build
WORKDIR /go/src/github.com/jhawk7/rpi-thermometer/
COPY . ./
RUN go mod download
RUN GOOS=linux GOARCH=arm go build -o thermo

FROM balenalib/rpi-raspbian:bullseye
WORKDIR /
COPY --from=build /go/src/github.com/jhawk7/rpi-thermometer/thermo ./
EXPOSE 8080
CMD ["./thermo"]
# i2c bus folder must be mounted from pi
