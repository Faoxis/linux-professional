package ru.otus.web;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.http.MediaType;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.RequestHeader;
import org.springframework.web.bind.annotation.RestController;

import java.net.InetAddress;
import java.time.OffsetDateTime;
import java.time.format.DateTimeFormatter;
import java.util.LinkedHashMap;
import java.util.Map;

@SpringBootApplication
@RestController
public class Application {

    private static final String APP = "Java Spring Boot";
    private static final String COLOR = "#e8590c";

    public static void main(String[] args) {
        SpringApplication.run(Application.class, args);
    }

    /** Данные, которые показываем на странице и отдаём в /api. */
    private Map<String, String> collect(String servedBy, String realIp) {
        Map<String, String> d = new LinkedHashMap<>();
        d.put("app", APP);
        d.put("runtime", "JVM " + System.getProperty("java.version")
                + " / Spring Boot " + SpringApplication.class.getPackage().getImplementationVersion());
        d.put("host", hostname());
        d.put("port", env("APP_PORT", "8080"));
        d.put("published", env("PUBLISHED_PORT", "?"));
        d.put("via", servedBy != null ? servedBy : "напрямую, без nginx");
        d.put("client", realIp != null ? realIp : "неизвестен");
        d.put("now", OffsetDateTime.now().format(DateTimeFormatter.ISO_OFFSET_DATE_TIME));
        return d;
    }

    private static String hostname() {
        try {
            return InetAddress.getLocalHost().getHostName();
        } catch (Exception e) {
            return "неизвестен";
        }
    }

    private static String env(String key, String def) {
        String v = System.getenv(key);
        return (v == null || v.isBlank()) ? def : v;
    }

    @GetMapping(value = "/", produces = MediaType.TEXT_HTML_VALUE + ";charset=UTF-8")
    public String index(@RequestHeader(value = "X-Served-By", required = false) String servedBy,
                        @RequestHeader(value = "X-Real-IP", required = false) String realIp) {
        Map<String, String> d = collect(servedBy, realIp);
        StringBuilder rows = new StringBuilder();
        row(rows, "рантайм", d.get("runtime"));
        row(rows, "контейнер", d.get("host"));
        row(rows, "слушает внутри сети", ":" + d.get("port"));
        row(rows, "опубликован на хосте", "localhost:" + d.get("published"));
        row(rows, "проксировал", d.get("via"));
        row(rows, "адрес клиента", d.get("client"));
        row(rows, "время ответа", d.get("now"));

        return """
                <!doctype html>
                <html lang="ru"><head><meta charset="utf-8"><title>%s</title>
                <style>
                 body{font:16px/1.5 system-ui,sans-serif;margin:0;background:#f8f9fa;color:#212529}
                 .card{max-width:640px;margin:48px auto;background:#fff;border-radius:12px;
                       box-shadow:0 1px 3px rgba(0,0,0,.12);overflow:hidden}
                 h1{margin:0;padding:24px 28px;background:%s;color:#fff;font-size:28px}
                 table{width:100%%;border-collapse:collapse}
                 td{padding:12px 28px;border-top:1px solid #e9ecef}
                 td:first-child{color:#868e96;width:44%%}
                 code{background:#f1f3f5;padding:2px 6px;border-radius:4px}
                </style></head><body>
                <div class="card"><h1>%s</h1><table>%s</table></div>
                </body></html>
                """.formatted(APP, COLOR, APP, rows);
    }

    private static void row(StringBuilder sb, String k, String v) {
        sb.append("<tr><td>").append(k).append("</td><td><code>").append(v).append("</code></td></tr>");
    }

    @GetMapping(value = "/api", produces = MediaType.APPLICATION_JSON_VALUE)
    public Map<String, String> api(@RequestHeader(value = "X-Served-By", required = false) String servedBy,
                                   @RequestHeader(value = "X-Real-IP", required = false) String realIp) {
        return collect(servedBy, realIp);
    }

    @GetMapping(value = "/health", produces = MediaType.TEXT_PLAIN_VALUE)
    public String health() {
        return "ok";
    }
}
