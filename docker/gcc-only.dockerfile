# Stage 1: Generate front-end
FROM node:18-alpine AS frontend

# Install node-gyp requirements
RUN apk add --no-cache python3 make g++

WORKDIR /kjudge/frontend

COPY ./frontend/package.json ./frontend/yarn.lock .
RUN yarn install --frozen-lockfile

COPY ./ /kjudge
RUN yarn --prod --frozen-lockfile build 

# Stage 2: Build back-end
FROM golang:alpine AS backend

RUN apk add --no-cache grep gcc g++ musl

WORKDIR /kjudge

COPY go.mod go.sum ./
RUN go mod download

COPY --from=frontend /kjudge/. /kjudge

RUN sh scripts/install_tools.sh

RUN go generate && go build -tags production -o kjudge cmd/kjudge/main.go

# Stage 3: Create awesome output image
FROM ghcr.io/minhnhatnoe/isolate:v2.1.4-alpine

RUN apk add --no-cache libcap make g++ openssl bash

COPY --from=backend /kjudge/kjudge /usr/local/bin
COPY --from=backend /kjudge/scripts /scripts

VOLUME ["/data", "/certs"]

EXPOSE 80 443

WORKDIR /
ENTRYPOINT ["scripts/start_container.sh"]
