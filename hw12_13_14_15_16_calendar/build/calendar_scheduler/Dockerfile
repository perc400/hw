# Build
FROM golang:1.23.11 AS build

ENV BIN_FILE /opt/calendar_scheduler/calendar-scheduler-service
ENV CODE_DIR /go/src/

WORKDIR ${CODE_DIR}

COPY go.mod go.sum ./
RUN go mod download

COPY . ${CODE_DIR}
ENV CGO_ENABLED=0 GOOS=linux GOARCH=amd64
RUN go build -o ${BIN_FILE} ./cmd/calendar_scheduler/*

# Runtime
FROM alpine:3.23.2

LABEL SERVICE="calendar"
ENV BIN_FILE /opt/calendar_scheduler/calendar-scheduler-service

COPY --from=build ${BIN_FILE} ${BIN_FILE}

ENTRYPOINT ["/opt/calendar_scheduler/calendar-scheduler-service"]
CMD ["-config", "/opt/calendar_scheduler/config.yaml"]
