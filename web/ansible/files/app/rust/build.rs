// Кладём версию компилятора в переменную окружения на этапе сборки,
// чтобы приложение могло показать её на странице.
use std::process::Command;

fn main() {
    let version = Command::new("rustc")
        .arg("--version")
        .output()
        .ok()
        .and_then(|o| String::from_utf8(o.stdout).ok())
        .map(|s| s.trim().to_string())
        .unwrap_or_else(|| "rustc (версия неизвестна)".to_string());

    println!("cargo:rustc-env=RUSTC_VERSION={version}");
}
