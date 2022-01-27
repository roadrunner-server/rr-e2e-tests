FROM ghcr.io/spiral/roadrunner:2.7.3 AS roadrunner
FROM php:8.1-cli
RUN apt update -y
RUN apt install build-essential -y

COPY --from=roadrunner /usr/bin/rr /usr/local/bin/rr

ENTRYPOINT ["/usr/local/bin/rr"]
