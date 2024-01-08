FROM debian:stretch-slim

WORKDIR /

COPY . /usr/local/bin

CMD ["custom-scheduler"]