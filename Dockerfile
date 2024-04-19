FROM golang:latest
RUN git pull

RUN go build -o output/ddns

RUN ./ddns