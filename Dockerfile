FROM golang:1.25.1 as builder

ARG CGO_ENABLED=0

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download
COPY . .

RUN go build -o access ./cmd/main.go 

FROM scratch
COPY --from=builder /app/access /access
COPY ./configs ./configs
EXPOSE 9083 9013

ENTRYPOINT ["/access"]
CMD ["--config-def-roles", "configs/role-config.json"]