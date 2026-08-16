//! Демонстрационное приложение на Rust: axum поверх tokio.

use axum::{
    extract::Request,
    http::{HeaderMap, StatusCode},
    response::{Html, IntoResponse},
    routing::get,
    Json, Router,
};
use serde_json::json;
use std::net::SocketAddr;

const APP: &str = "Rust";
const COLOR: &str = "#dea584"; // цвет Rust на github-linguist

/// Данные, которые показываем на странице и отдаём в /api.
fn collect(headers: &HeaderMap) -> serde_json::Value {
    json!({
        "app": APP,
        // RUSTC_VERSION подставляет build.rs на этапе сборки
        "runtime": format!("{} / axum 0.7", env!("RUSTC_VERSION")),
        "host": hostname(),
        "port": env("APP_PORT", "8080"),
        "published": env("PUBLISHED_PORT", "?"),
        "via": hdr(headers, "x-served-by", "напрямую, без nginx"),
        "client": hdr(headers, "x-real-ip", "неизвестен"),
        "now": now(),
    })
}

fn env(key: &str, default: &str) -> String {
    std::env::var(key)
        .ok()
        .filter(|v| !v.is_empty())
        .unwrap_or_else(|| default.to_string())
}

/// Значение заголовка или подстановка по умолчанию.
/// Названа hdr, а не header: имя header заняло бы модуль axum::http::header.
fn hdr(headers: &HeaderMap, name: &str, default: &str) -> String {
    headers
        .get(name)
        .and_then(|v| v.to_str().ok())
        .filter(|v| !v.is_empty())
        .unwrap_or(default)
        .to_string()
}

fn hostname() -> String {
    std::fs::read_to_string("/etc/hostname")
        .map(|s| s.trim().to_string())
        .unwrap_or_else(|_| "неизвестен".to_string())
}

/// Время в формате RFC 3339. Отдельного крейта ради этого не тянем:
/// считаем от эпохи вручную.
fn now() -> String {
    use std::time::{SystemTime, UNIX_EPOCH};
    let secs = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|d| d.as_secs())
        .unwrap_or(0);
    let (days, rem) = (secs / 86_400, secs % 86_400);
    let (h, m, s) = (rem / 3600, (rem % 3600) / 60, rem % 60);
    let (y, mo, d) = civil_from_days(days as i64);
    format!("{y:04}-{mo:02}-{d:02}T{h:02}:{m:02}:{s:02}Z")
}

/// Перевод числа суток от эпохи в календарную дату,
/// алгоритм Говарда Хиннанта.
fn civil_from_days(z: i64) -> (i64, u32, u32) {
    let z = z + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z.rem_euclid(146_097);
    let yoe = (doe - doe / 1460 + doe / 36_524 - doe / 146_096) / 365;
    let y = yoe + era * 400;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = (doy - (153 * mp + 2) / 5 + 1) as u32;
    let m = if mp < 10 { mp + 3 } else { mp - 9 } as u32;
    (if m <= 2 { y + 1 } else { y }, m, d)
}

async fn index(req: Request) -> impl IntoResponse {
    let d = collect(req.headers());
    let get = |k: &str| d[k].as_str().unwrap_or("").to_string();
    let row = |k: &str, v: &str| format!("<tr><td>{k}</td><td><code>{v}</code></td></tr>");

    let rows = [
        row("рантайм", &get("runtime")),
        row("контейнер", &get("host")),
        row("слушает внутри сети", &format!(":{}", get("port"))),
        row("опубликован на хосте", &format!("localhost:{}", get("published"))),
        row("проксировал", &get("via")),
        row("адрес клиента", &get("client")),
        row("время ответа", &get("now")),
    ]
    .join("");

    Html(format!(
        r#"<!doctype html>
<html lang="ru"><head><meta charset="utf-8"><title>{APP}</title>
<style>
 body{{font:16px/1.5 system-ui,sans-serif;margin:0;background:#f8f9fa;color:#212529}}
 .card{{max-width:640px;margin:48px auto;background:#fff;border-radius:12px;
       box-shadow:0 1px 3px rgba(0,0,0,.12);overflow:hidden}}
 h1{{margin:0;padding:24px 28px;background:{COLOR};color:#fff;font-size:28px}}
 table{{width:100%;border-collapse:collapse}}
 td{{padding:12px 28px;border-top:1px solid #e9ecef}}
 td:first-child{{color:#868e96;width:44%}}
 code{{background:#f1f3f5;padding:2px 6px;border-radius:4px}}
</style></head><body>
<div class="card"><h1>{APP}</h1><table>{rows}</table></div>
</body></html>"#
    ))
}

async fn api(req: Request) -> impl IntoResponse {
    Json(collect(req.headers()))
}

async fn health() -> impl IntoResponse {
    ([("content-type", "text/plain; charset=utf-8")], "ok")
}

#[tokio::main]
async fn main() {
    let port: u16 = env("APP_PORT", "8080").parse().unwrap_or(8080);

    let app = Router::new()
        .route("/", get(index))
        .route("/api", get(api))
        .route("/health", get(health))
        .fallback(|| async { (StatusCode::NOT_FOUND, "not found") });

    let addr = SocketAddr::from(([0, 0, 0, 0], port));
    println!("{APP} слушает {addr}");
    let listener = tokio::net::TcpListener::bind(addr).await.unwrap();
    axum::serve(listener, app).await.unwrap();
}
