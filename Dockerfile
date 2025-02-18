FROM node:22-alpine AS builder-web

WORKDIR /app
COPY ./ui/package.json ./ui/yarn.lock .
RUN yarn --frozen-lockfile

COPY ./ui .

RUN yarn build

FROM golang:1.23-alpine AS builder

WORKDIR /app
COPY ./go.mod ./go.sum .

RUN apk add --no-cache make build-base proj-dev
RUN go mod download

COPY . .
COPY --from=builder-web /app/dist ./ui/dist

RUN go build -ldflags="-s -w" -o treeregister-plugin 

FROM alpine:3.20 AS runner

RUN adduser -D gorunner
RUN apk add --no-cache proj

USER gorunner

WORKDIR /app

COPY --chown=gorunner:gorunner --from=builder /app/treeregister-plugin .

ENTRYPOINT [ "/app/treeregister-plugin" ]




