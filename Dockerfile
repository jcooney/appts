FROM golang:1.25.1-alpine AS builder

# Create app directory
WORKDIR /workspace
COPY . .

# Build app
RUN --mount=type=cache,target=/root/.cache/go-build go build -tags timetzdata -o dist/api cmd/main.go

FROM alpine:latest
WORKDIR /app
COPY --from=builder /workspace/dist/api ./
COPY --from=builder /workspace/schema/ddl ./schema/ddl

CMD [ "./api" ]
EXPOSE 3333