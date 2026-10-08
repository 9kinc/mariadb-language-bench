FROM php:8.5-fpm-bookworm
RUN docker-php-ext-install -j2 pdo_mysql
COPY php/zz-bench.conf /usr/local/etc/php-fpm.d/zz-bench.conf
COPY php/index.php /var/www/html/index.php
