# Stage 0: Compile isolate
FROM alpine:3 AS isolate

RUN apk add --no-cache libcap gcc make git g++ libcap-dev

WORKDIR /isolate

RUN git clone --branch v1.8.1 --single-branch https://github.com/ioi/isolate.git .

RUN make isolate

# Stage 1: Generate front-end
FROM node:13-alpine AS frontend

COPY ./ /kjudge

WORKDIR /kjudge/frontend

RUN yarn && yarn --prod --frozen-lockfile build 

# Stage 3: Build back-end
FROM golang:alpine AS backend

RUN apk add --no-cache grep gcc g++ musl

WORKDIR /kjudge

COPY go.mod go.sum ./
RUN go mod download

COPY --from=frontend /kjudge/. /kjudge

RUN sh scripts/install_tools.sh 
RUN sed -i 's/^debug/# debug/' fileb0x.yaml
RUN go generate && go build -tags production -o kjudge cmd/kjudge/main.go

# Stage 4: Create awesome output image
FROM alpine:3

RUN apk add --no-cache libcap make g++

COPY --from=isolate /isolate/ /isolate

WORKDIR /isolate
RUN make install

COPY --from=backend /kjudge/kjudge /usr/local/bin

VOLUME ["/data"]

EXPOSE 80

CMD ["kjudge", "-port", "80", "-file", "/data/kjudge.db"]
