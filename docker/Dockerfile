
FROM golang:1.20

# install whatap-agent (x64)
RUN wget https://s3.ap-northeast-2.amazonaws.com/repo.whatap.io/alpine/x86_64/whatap-agent.tar.gz
RUN tar -xvzf whatap-agent.tar.gz -C /

# Create whatap.conf. You can copy it or attach it as a volume.
RUN echo "accesskey=aaasd-23432-123-" >> /app/whatap.conf
RUN echo "whatap.server.host=1.1.1.1/2.2.2.2" >> /app/whatap.conf

# Whatap/go-api library must be imported into the user application.
WORKDIR /app
RUN go mod tidy
RUN go mod download -x
RUN go build -o ./app app.go

# Run whatap-agent and the user application together with the Docker execution command.
CMD ['sh', '-c', '/etc/init.d/whatap-agent start && /app/app']