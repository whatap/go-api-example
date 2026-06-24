
FROM golang:1.20

# install whatap-agent (x64)
COPY --from=whatap-go-inst-builder /usr/whatap/agent/ /usr/whatap/agent/

# Create whatap.conf. You can copy it or attach it as a volume.
RUN echo "accesskey=aaasd-23432-123-" >> /app/whatap.conf
RUN echo "whatap.server.host=1.1.1.1/2.2.2.2" >> /app/whatap.conf

# Whatap/go-api library must be imported into the user application.
WORKDIR /app
ENV WHATAP_HOME /app

RUN go mod tidy
RUN go mod download -x
RUN go build -o ./app app.go

# Run whatap-agent and the user application together with the Docker execution command.
CMD ['sh', '-c', '/usr/whatap/agent/whatap-agent start && /app/app']