#!/bin/bash
#
# log_report.sh — анализ nginx access log и отправка отчёта на email
#
# Использование:
#   ./log_report.sh [LOG_FILE] [EMAIL]
#
# Cron (каждый час):
#   0 * * * * /path/to/log_report.sh /path/to/access.log user@example.com
#

set -euo pipefail

# ===== Конфигурация =====
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
LOG_FILE="${1:-${SCRIPT_DIR}/access-4560-644067.log}"
EMAIL="${2:-root@localhost}"
LOCK_FILE="/tmp/log_report.lock"
STATE_FILE="/tmp/log_report_last_line"
TOP_N=20

# ===== Проверка лог-файла =====
if [[ ! -f "$LOG_FILE" ]]; then
    echo "Лог-файл не найден: ${LOG_FILE}" >&2
    exit 1
fi

# ===== Защита от параллельного запуска =====
if command -v flock &>/dev/null; then
    exec 200>"$LOCK_FILE"
    if ! flock -n 200; then
        echo "Скрипт уже запущен. Выход." >&2
        exit 1
    fi
else
    if [[ -f "$LOCK_FILE" ]]; then
        OLD_PID=$(cat "$LOCK_FILE" 2>/dev/null || true)
        if [[ -n "$OLD_PID" ]] && kill -0 "$OLD_PID" 2>/dev/null; then
            echo "Скрипт уже запущен (PID: ${OLD_PID}). Выход." >&2
            exit 1
        fi
        rm -f "$LOCK_FILE"
    fi
    echo $$ > "$LOCK_FILE"
    LOCK_VIA_PID=1
fi

# ===== Определение диапазона строк =====
TOTAL_LINES=$(wc -l < "$LOG_FILE" | tr -d ' ')

if [[ -f "$STATE_FILE" ]]; then
    LAST_LINE=$(cat "$STATE_FILE")
else
    LAST_LINE=0
fi

if (( LAST_LINE >= TOTAL_LINES )); then
    echo "Нет новых строк для обработки."
    exit 0
fi

START_LINE=$(( LAST_LINE + 1 ))
END_LINE=$TOTAL_LINES

# ===== Извлечение новых строк во временный файл =====
TMP_FILE=$(mktemp)
trap 'rm -f "$TMP_FILE"; [[ "${LOCK_VIA_PID:-}" == 1 ]] && rm -f "$LOCK_FILE"' EXIT

sed -n "${START_LINE},${END_LINE}p" "$LOG_FILE" | sed '1s/^Z //' > "$TMP_FILE"

# ===== Временной диапазон =====
TIME_FROM=$(head -1 "$TMP_FILE" | awk -F'[][]' '{print $2}')
TIME_TO=$(tail -1 "$TMP_FILE" | awk -F'[][]' '{print $2}')

# ===== Анализ =====

TOP_IP=$(awk '{print $1}' "$TMP_FILE" | sort | uniq -c | sort -rn | head -"$TOP_N")

TOP_URL=$(awk '{print $7}' "$TMP_FILE" | sort | uniq -c | sort -rn | head -"$TOP_N")

HTTP_CODES=$(awk '{print $9}' "$TMP_FILE" | sort | uniq -c | sort -rn)

ERRORS=$(awk '$9 ~ /^[45][0-9][0-9]$/ {print $9, $7}' "$TMP_FILE" \
    | sort | uniq -c | sort -rn | head -"$TOP_N")

# ===== Формирование отчёта =====
REPORT="===== Отчёт по access-логу =====
Временной диапазон: ${TIME_FROM} — ${TIME_TO}
Обработано строк: ${START_LINE}–${END_LINE} ($(( END_LINE - START_LINE + 1 )) новых)

--- Top ${TOP_N} IP-адресов по числу запросов ---
${TOP_IP}

--- Top ${TOP_N} запрашиваемых URL ---
${TOP_URL}

--- HTTP-коды ответов (все) ---
${HTTP_CODES}

--- Ошибки веб-сервера/приложения (4xx/5xx): код + URL (top ${TOP_N}) ---
${ERRORS}

===== Конец отчёта ====="

# ===== Отправка / вывод =====
if command -v mail &>/dev/null; then
    echo "$REPORT" | mail -s "Nginx log report: ${TIME_FROM} — ${TIME_TO}" "$EMAIL"
    echo "Отчёт отправлен на ${EMAIL}"
else
    echo "$REPORT"
    echo ""
    echo "(утилита mail не найдена — отчёт выведен в stdout)"
fi

# ===== Обновление состояния =====
echo "$END_LINE" > "$STATE_FILE"
