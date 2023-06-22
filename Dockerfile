FROM golang:1.17 as builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . ./
RUN CGO_ENABLED=0 GOOS=linux go build -v -o myapp

FROM alpine:latest
RUN apk --no-cache add ca-certificates python3 py3-pip
RUN pip3 install --upgrade pip
RUN pip3 install awscli
ENV KUBECONFIG /etc/kubeconfig/config
COPY ./kubeconfig /etc/kubeconfig/config
COPY --from=builder /app/myapp /myapp
CMD ["/myapp"]
