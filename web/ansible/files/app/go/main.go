// Демонстрационное приложение на Go. Только стандартная библиотека,
// внешних зависимостей нет.
package main

import (
	"encoding/json"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"os"
	"runtime"
	"time"
)

const appName = "Go"

// Страница одинаковая по верстке у всех трёх приложений, меняются только данные.
var page = template.Must(template.New("page").Parse(`<!doctype html>
<html lang="ru"><head><meta charset="utf-8"><title>{{.App}}</title>
<style>
 body{font:16px/1.5 system-ui,sans-serif;margin:0;background:#f8f9fa;color:#212529}
 .card{max-width:640px;margin:48px auto;background:#fff;border-radius:12px;
       box-shadow:0 1px 3px rgba(0,0,0,.12);overflow:hidden}
 h1{margin:0;padding:24px 28px;background:{{.Color}};color:#fff;font-size:28px}
 table{width:100%;border-collapse:collapse}
 td{padding:12px 28px;border-top:1px solid #e9ecef}
 td:first-child{color:#868e96;width:44%}
 code{background:#f1f3f5;padding:2px 6px;border-radius:4px}
</style></head><body>
<div class="card">
  <h1>{{.App}}</h1>
  <table>
    <tr><td>рантайм</td><td><code>{{.Runtime}}</code></td></tr>
    <tr><td>контейнер</td><td><code>{{.Host}}</code></td></tr>
    <tr><td>слушает внутри сети</td><td><code>:{{.Port}}</code></td></tr>
    <tr><td>опубликован на хосте</td><td><code>localhost:{{.Published}}</code></td></tr>
    <tr><td>проксировал</td><td><code>{{.Via}}</code></td></tr>
    <tr><td>адрес клиента</td><td><code>{{.Client}}</code></td></tr>
    <tr><td>время ответа</td><td><code>{{.Now}}</code></td></tr>
  </table>
</div></body></html>`))

// Теги нужны, чтобы /api отдавал такие же ключи, как остальные два
// приложения. Color - деталь оформления страницы, в JSON ему не место.
type data struct {
	App       string `json:"app"`
	Runtime   string `json:"runtime"`
	Host      string `json:"host"`
	Port      string `json:"port"`
	Published string `json:"published"`
	Via       string `json:"via"`
	Client    string `json:"client"`
	Now       string `json:"now"`
	Color     string `json:"-"`
}

func collect(r *http.Request) data {
	host, _ := os.Hostname()
	return data{
		App:       appName,
		Runtime:   runtime.Version(),
		Host:      host,
		Port:      env("APP_PORT", "8080"),
		Published: env("PUBLISHED_PORT", "?"),
		Via:       header(r, "X-Served-By", "напрямую, без nginx"),
		Client:    header(r, "X-Real-IP", r.RemoteAddr),
		Now:       time.Now().Format(time.RFC3339),
		Color:     "#00add8", // фирменный цвет Go
	}
}

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func header(r *http.Request, k, def string) string {
	if v := r.Header.Get(k); v != "" {
		return v
	}
	return def
}

func main() {
	port := env("APP_PORT", "8080")

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		if err := page.Execute(w, collect(r)); err != nil {
			log.Println("шаблон:", err)
		}
	})

	http.HandleFunc("/api", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json; charset=utf-8")
		enc := json.NewEncoder(w)
		// По умолчанию encoding/json экранирует < > & в \uXXXX на случай
		// вставки JSON в html. Мы отдаём его как api, экранирование только мешает.
		enc.SetEscapeHTML(false)
		enc.Encode(collect(r))
	})

	http.HandleFunc("/health", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "ok")
	})

	log.Printf("%s слушает :%s", appName, port)
	log.Fatal(http.ListenAndServe(":"+port, nil))
}
