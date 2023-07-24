## Build
FROM golang:1.20 AS build

# backend 
WORKDIR /app

COPY go.* ./
RUN go mod download

COPY . ./

#RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X main.compileDate=`date -u +.%Y%m%d.%H%M%S`" -o /sistema-de-rifas .

RUN CGO_ENABLED=1 go build -o /forum .

# frontend
#FROM gcr.io/distroless/static-debian11
FROM frolvlad/alpine-glibc

WORKDIR /

# executável principal
COPY --from=build /forum /forum

EXPOSE 3000

CMD ["tail", "-f", "/dev/null"]
