#!/bin/bash
#
# Снятие резервной копии на удалённый borg-репозиторий.
# Запускается юнитом borg-backup.service из таймера borg-backup.timer.
# Реквизиты приходят из /etc/borg/borg.env через EnvironmentFile,
# весь вывод юнит заворачивает в logger с тегом borg-backup.

set -o pipefail

: "${BORG_REPO:?не задан BORG_REPO}"
: "${BACKUP_TARGET:=/etc}"
: "${KEEP_DAILY:=90}"
: "${KEEP_MONTHLY:=12}"

# Имя архива содержит время снятия. Фигурные скобки раскрывает сам borg,
# поэтому строка в одинарных кавычках и не трогается шеллом.
ARCHIVE='etc-{now:%Y-%m-%d_%H:%M:%S}'

rc=0
step() { echo "--- $*"; }
fail() { echo "ОШИБКА на шаге '$1', код $2"; rc=$2; }

step "старт: цель ${BACKUP_TARGET}, репозиторий ${BORG_REPO}"

# --stats печатает объём и дедупликацию, --show-rc - код возврата borg
step "создание архива"
borg create --stats --show-rc --compression lz4 \
    "::${ARCHIVE}" "${BACKUP_TARGET}" \
    || fail "borg create" $?

# Политика хранения. --glob-archives ограничивает выборку нашими архивами,
# чтобы prune не тронул чужие, если в репозиторий начнёт писать кто-то ещё.
step "политика хранения: --keep-daily ${KEEP_DAILY} --keep-monthly ${KEEP_MONTHLY}"
borg prune --list --show-rc \
    --glob-archives 'etc-*' \
    --keep-daily "${KEEP_DAILY}" \
    --keep-monthly "${KEEP_MONTHLY}" \
    || fail "borg prune" $?

# В borg 1.2 prune только помечает данные ненужными, место освобождает compact
step "освобождение места"
borg compact --show-rc || fail "borg compact" $?

# Проверка целостности репозитория. На боевом объёме это долго и её выносят
# в отдельный редкий таймер, здесь репозиторий крошечный.
step "проверка репозитория"
borg check --show-rc || fail "borg check" $?

step "готово, итоговый код ${rc}"
exit "${rc}"
