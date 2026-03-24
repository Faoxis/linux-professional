FROM alpine:latest

LABEL org.opencontainers.image.title="otus_my_homework"

RUN apk add --no-cache nginx

COPY default.conf /etc/nginx/http.d/default.conf
COPY index.html /usr/share/nginx/html/index.html

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
