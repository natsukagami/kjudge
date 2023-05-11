# Stage 0: Compile isolate
FROM ubuntu:jammy AS isolate

RUN apt-get update && apt-get install -y libcap-dev gcc git make

WORKDIR /isolate

RUN git clone --branch v1.10.1 --single-branch https://github.com/ioi/isolate.git .

RUN make isolate

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
FROM ubuntu:jammy

RUN apt-get update && DEBIAN_FRONTEND="noninteractive" TZ="Asia/Ho_Chi_Minh" apt-get install -y \
    build-essential openjdk-11-jdk-headless fp-compiler python3.6 cgroup-lite python2.7 rustc golang libcap-dev # Compilers  \
    openssl # For HTTPS support

COPY --from=isolate /isolate/ /isolate

WORKDIR /isolate
RUN make install

COPY --from=backend /kjudge/kjudge /usr/local/bin
COPY --from=backend /kjudge/scripts /scripts

RUN ln -s /usr/bin/python2.7 /usr/bin/python2

VOLUME ["/data", "/certs"]

EXPOSE 80 443

WORKDIR /
ENTRYPOINT ["scripts/start_container.sh"]

