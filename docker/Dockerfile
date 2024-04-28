# Stage 0: Update apt-get
FROM ubuntu:jammy AS base-ubuntu

# Mount apt's cache folders to cache install but maintain
# reproducibility on local machines. More details at:
# https://github.com/moby/buildkit/blob/master/frontend/dockerfile/docs/reference.md#run---mounttypecache

# Set sharing=locked to avoid race condition with between stages.
RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && apt-get upgrade -y

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
FROM golang:1.18-bullseye AS backend

WORKDIR /kjudge

COPY go.mod go.sum ./
RUN go mod download

COPY --from=frontend /kjudge/. /kjudge

RUN scripts/install_tools.sh
RUN go generate && go build -tags production -o kjudge cmd/kjudge/main.go

# Stage 3: Create awesome output image
FROM ghcr.io/minhnhatnoe/isolate:latest

RUN --mount=type=cache,target=/var/cache/apt,sharing=locked \
    --mount=type=cache,target=/var/lib/apt,sharing=locked \
    apt-get update && DEBIAN_FRONTEND="noninteractive" TZ="Asia/Ho_Chi_Minh" apt-get install -y \
    build-essential openjdk-11-jdk-headless fp-compiler cgroup-lite \
    python3.10 python2.7 rustc golang libcap-dev \
    openssl

COPY --from=backend /kjudge/kjudge /usr/local/bin
COPY --from=backend /kjudge/scripts /scripts

RUN ln -s /usr/bin/python2.7 /usr/bin/python2

VOLUME ["/data", "/certs"]

EXPOSE 80 443

WORKDIR /
ENTRYPOINT ["scripts/start_container.sh"]
