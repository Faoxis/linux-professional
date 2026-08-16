# Многоступенчатая сборка: компилируем в образе с тулчейном,
# в финальный образ кладём только один бинарник.
FROM golang:1.22-alpine AS build
WORKDIR /src
COPY go.mod ./
COPY main.go ./
# CGO выключен, иначе бинарник потянет за собой libc из образа сборки
RUN CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o /app .

FROM alpine:3.20
RUN adduser -D -u 10001 app
COPY --from=build /app /usr/local/bin/app
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/app"]
