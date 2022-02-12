FROM ghcr.io/roadrunner-server/roadrunner:2.7.7 AS roadrunner
FROM php:7.4-cli

COPY --from=roadrunner /usr/bin/rr /usr/local/bin/rr

RUN mkdir "/src"
WORKDIR "/src"

RUN apt update -y
RUN apt upgrade -y

RUN docker-php-ext-install sockets
RUN apt install build-essential -y
RUN apt install git zip unzip -y

COPY .rr.yaml .
COPY psr-worker-bench.php .
COPY composer.json .

RUN curl -sS https://getcomposer.org/installer | php -- --install-dir=/usr/local/bin --filename=composer
RUN composer update
EXPOSE 15389

ENTRYPOINT ["/usr/local/bin/rr"]
