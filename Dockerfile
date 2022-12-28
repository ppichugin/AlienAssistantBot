################################
# STEP 1 build executable binary
################################
FROM golang:alpine AS build
# Install git + SSL ca certificates (to support https requests to Telegram API).
# Git is required for fetching the dependencies.
# Ca-certificates is required to call HTTPS endpoints.
RUN apk update && apk add --no-cache git ca-certificates tzdata && update-ca-certificates
# Create appuser
ENV USER=appuser
ENV UID=10001
# See https://stackoverflow.com/a/55757473/12429735RUN
RUN adduser \
    --disabled-password \
    --gecos "" \
    --home "/nonexistent" \
    --shell "/sbin/nologin" \
    --no-create-home \
    --uid "${UID}" \
    "${USER}"

WORKDIR /go/src/github.com/ppichugin/AlienAssistantBot/
COPY . .
RUN go get -d -v
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -o alien-assitant-bot .

############################
# STEP 2 build a small image
############################
FROM scratch
# Set some image labels
LABEL author="petr.pichugin@gmail.com"
LABEL version="1.0"
LABEL description="Telegram-bot backend software packed to docker image. \
Telegram-bot link: https://t.me/AlienAssistantBot"
# Import the user and group files from the builder.
COPY --from=build /etc/passwd /etc/passwd
COPY --from=build /etc/group /etc/group
# Copy certificates
COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
# Copy zoneinfo for timezones
COPY --from=build /usr/share/zoneinfo /usr/share/zoneinfo
# Copy static executable
COPY --from=build /go/src/github.com/ppichugin/AlienAssistantBot/ ./
# Use an unprivileged user
USER appuser:appuser
# Run binary
CMD ["./alien-assitant-bot"]