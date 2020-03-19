# Stage 0: Compile isolate
FROM ubuntu:bionic

RUN apt-get update && apt-get install -y libcap-dev gcc git make

WORKDIR /isolate

RUN git clone --branch v1.8.1 --single-branch https://github.com/ioi/isolate.git .

RUN make isolate

# Stage 1: git clone 
FROM ubuntu:bionic

WORKDIR /kjudge

RUN apt-get update && apt-get install -y git

RUN git clone --branch v0.1.1 https://git.nkagami.me/natsukagami/kjudge --single-branch .

# Stage 2: Generate front-end
FROM node:13

COPY --from=1 /kjudge/ /kjudge

WORKDIR /kjudge/frontend

RUN yarn && yarn --prod --frozen-lockfile build 

# Stage 3: Build back-end
FROM golang:1.13-stretch

COPY --from=2 /kjudge/ /kjudge

WORKDIR /kjudge

RUN go mod download
RUN scripts/install_tools.sh 
RUN sed -i 's/^debug/# debug/' fileb0x.yaml
RUN go generate && go build -tags production -o kjudge cmd/kjudge/main.go

# Stage 4: Create awesome output image
FROM ubuntu:bionic

RUN apt-get update && apt-get install -y build-essential openjdk-8-jdk-headless fp-compiler python3.6 cgroup-lite python2.7 rustc golang libcap-dev

COPY --from=0 /isolate/ /isolate

WORKDIR /isolate
RUN make install

COPY --from=3 /kjudge/kjudge /usr/local/bin

RUN ln -s /usr/bin/python3.6 /usr/bin/python3 && ln -s /usr/bin/python2.7 /usr/bin/python2

VOLUME ["/data"]

EXPOSE 80

CMD ["kjudge", "-port", "80", "-file", "/data/kjudge.db"]
