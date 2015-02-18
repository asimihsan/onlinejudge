#!/bin/bash

set -e

sudo tee /etc/apt/sources.list.d/nginx-stable-trusty.list >/dev/null <<EOF
deb http://ppa.launchpad.net/nginx/stable/ubuntu trusty main
EOF

sudo apt-key adv --keyserver keyserver.ubuntu.com --recv-keys C300EE8C
sudo apt-get --assume-yes --quiet update
sudo apt-get --assume-yes --quiet install nginx

sudo tee /etc/nginx/nginx.conf >/dev/null <<EOF
user www-data;
worker_processes 1;  # good for one core
worker_priority 15;
pid /run/nginx.pid;

events {
    worker_connections 1024;  # good for 512MB host
    multi_accept on;
    use epoll;
}

http {
    sendfile on;
    tcp_nopush on;
    tcp_nodelay on;

    client_body_timeout 12;
    client_header_timeout 12;
    keepalive_timeout 15;
    send_timeout 10;

    client_body_buffer_size 10K;
    client_header_buffer_size 1k;
    client_max_body_size 8m;
    large_client_header_buffers 2 1k;

    types_hash_max_size 2048;
    include /etc/nginx/mime.types;
    default_type application/octet-stream;
    access_log /var/log/nginx/access.log;
    error_log /var/log/nginx/error.log;

    gzip on;
    gzip_disable "msie6";
    gzip_vary on;
    gzip_proxied expired no-cache no-store private auth;
    gzip_comp_level 6;
    gzip_buffers 16 8k;
    gzip_http_version 1.1;
    gzip_types text/plain text/css application/json application/javascript application/x-javascript text/xml application/xml application/xml+rss text/javascript;

    include /etc/nginx/conf.d/*.conf;
    include /etc/nginx/sites-enabled/*;
}
EOF

# need to quote heredoc start word, or else dollar signs
# not escaped in host shell
sudo tee /etc/nginx/sites-enabled/default >/dev/null <<"EOF"
server {
    server_name www.runsomecode.com;
    listen 80;
    return 301 $scheme://runsomecode.com$request_uri;
}

server {
    server_name runsomecode.com;
    listen 80 default_server;
    listen [::]:80 default_server ipv6only=on;
    root /usr/share/nginx/html;
    index index.html index.htm;
    location ^~ /run {
        proxy_pass http://localhost:8080;
    }
    location ^~ /evaluator {
        proxy_pass http://localhost:8081;
    }
    location ^~ /auth {
        proxy_pass http://localhost:9001;
    }
    location / {
        try_files $uri $uri/ =404;
    }
}
EOF

sudo service nginx restart
