# Многоступенчатая сборка: тулчейн Rust весит больше гигабайта,
# в финальный образ кладём только собранный бинарник.
FROM rust:1-slim-bookworm AS build
WORKDIR /src
# Сначала манифест и заглушка main: слой с зависимостями переиспользуется,
# пока Cargo.toml не меняется
COPY Cargo.toml build.rs ./
RUN mkdir src && echo 'fn main(){}' > src/main.rs && \
    cargo build --release && rm -rf src
COPY src ./src
# Трогаем main.rs, иначе cargo сочтёт сборку актуальной по заглушке
RUN touch src/main.rs && cargo build --release

FROM debian:bookworm-slim
RUN useradd -u 10001 -m app
COPY --from=build /src/target/release/otus-web-rust /usr/local/bin/app
USER app
EXPOSE 8080
ENTRYPOINT ["/usr/local/bin/app"]
