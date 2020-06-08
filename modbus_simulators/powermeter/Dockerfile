# Use a 2-stage build with the final container as "scratch" to
# minimise final image size. As go creates a static binary, we
# only need copy the final executable to the scratch container.
FROM golang:1.14 as builder

# Create a new user so container is not run as root
RUN useradd pm
WORKDIR /powermeter

# Fetch the dependencies as a separate step to
# allow caching on each build.
COPY go.mod go.sum ./
RUN go mod download

# Build the executable
COPY main.go ./
RUN  CGO_ENABLED=0 go build

FROM scratch
# Copy across the user information from the builder
COPY --from=builder /etc/passwd /etc/passwd
COPY --from=builder /etc/group /etc/group

USER pm

COPY --from=builder /powermeter/powermeter .

ENTRYPOINT ["./powermeter"]