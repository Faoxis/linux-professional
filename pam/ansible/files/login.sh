#!/bin/bash
#
# Проверка права на вход, вызывается из PAM через pam_exec.
# Правило: в субботу и воскресенье пускаем только членов группы админов.
# Праздничный день из списка выходным не считается - в него пускаем всех.
#
# Коды возврата: 0 - пустить, 1 - не пускать.
# Имя пользователя PAM передаёт в переменной PAM_USER.

ADMIN_GROUP=admin
HOLIDAYS=/etc/pam-holidays
LOG=/var/log/pam-weekend.log

user="${PAM_USER:-$(id -un)}"
today=$(date +%F)
# %u: 1 - понедельник ... 6 - суббота, 7 - воскресенье
dow=$(date +%u)

decide() {
    echo "$(date '+%F %T') user=$user day=$today dow=$dow $1: $2" >> "$LOG"
    [ "$1" = allow ] && exit 0 || exit 1
}

# Членство проверяем через id, а не через /etc/group: в группе,
# записанной пользователю основной, его имени в /etc/group нет.
if id -nG "$user" 2>/dev/null | tr ' ' '\n' | grep -qx "$ADMIN_GROUP"; then
    decide allow "состоит в группе $ADMIN_GROUP"
fi

if [ "$dow" -lt 6 ]; then
    decide allow "будний день"
fi

# Выходной, но проверяем, не праздник ли. Комментарии из файла отбрасываем.
if [ -f "$HOLIDAYS" ] && grep -vE '^[[:space:]]*#' "$HOLIDAYS" |
        grep -qE "^[[:space:]]*${today}([[:space:]]|#|$)"; then
    decide allow "праздничный день"
fi

decide deny "выходной день, не состоит в группе $ADMIN_GROUP"
