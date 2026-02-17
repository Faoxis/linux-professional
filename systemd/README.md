# Домашнее задание: systemd 


## 1) watchlog: сервис, который раз в 30 секунд мониторит лог на ключевое слово

### 1.1 Конфигурация сервиса: `/etc/default/watchlog`

```console
root@new-vm:/home/faoxis# cat /etc/default/watchlog
# Configuration file for my watchlog service
# Place it to /etc/default

# File and word in that file that we will be monit
WORD="ALERT"
LOG=/var/log/watchlog.log
```

### 1.2 Тестовый лог: `/var/log/watchlog.log`

```console
root@new-vm:/home/faoxis# cat /var/log/watchlog.log
ALERT hey hey
ALERT my my
ALERT rock and roll will never die
```

### 1.3 Скрипт проверки: `/opt/watchlog.sh`

```bash
#!/bin/bash

WORD=$1
LOG=$2
DATE=`date`

if grep $WORD $LOG &> /dev/null
then
logger "$DATE: I found word, Master!"
else
exit 0
fi
```

Права на исполнение:

```console
root@new-vm:/home/faoxis# chmod +x /opt/watchlog.sh
```

### 1.4 Unit-файл сервиса: `/etc/systemd/system/watchlog.service`

```ini
[Unit]
Description=My watchlog service

[Service]
Type=oneshot
EnvironmentFile=/etc/default/watchlog
ExecStart=/opt/watchlog.sh $WORD $LOG
```

### 1.5 Unit-файл таймера: `/etc/systemd/system/watchlog.timer`

```ini
[Unit]
Description=Run watchlog script every 30 second

[Timer]
# Run every 30 second
OnUnitActiveSec=30
Unit=watchlog.service

[Install]
WantedBy=multi-user.target
```

### 1.6 Запуск и проверка

Запуск таймера:

```console
root@new-vm:/home/faoxis# systemctl start watchlog.timer
```

Запуск сервиса (это необходимо для активации сервиса, иначе `OnUnitActiveSec` не сработает):

```console
root@new-vm:/home/faoxis# systemctl start watchlog.service
```

Проверка, что событие логируется в syslog через `logger`:

```console
root@new-vm:/home/faoxis# tail -n 100 /var/log/syslog | grep "word"
2026-02-17T21:57:16.163183+00:00 new-vm root: Tue Feb 17 21:57:16 UTC 2026: I found word, Master!
2026-02-17T21:58:04.352212+00:00 new-vm root: Tue Feb 17 21:58:04 UTC 2026: I found word, Master!
2026-02-17T21:58:44.351274+00:00 new-vm root: Tue Feb 17 21:58:44 UTC 2026: I found word, Master!
```

---

## 2) spawn-fcgi: установка и systemd unit (переделка init-скрипта)

### 2.1 Установка пакетов

```console
root@new-vm:/home/faoxis# apt update
root@new-vm:/home/faoxis# apt install spawn-fcgi php php-cgi php-cli apache2 libapache2-mod-fcgid -y
```

### 2.2 Конфигурация spawn-fcgi: `/etc/spawn-fcgi/fcgi.conf`

```console
root@new-vm:/home/faoxis# cat /etc/spawn-fcgi/fcgi.conf
SOCKET=/var/run/php-fcgi.sock
OPTIONS="-u www-data -g www-data -s $SOCKET -S -M 0600 -C 32 -F 1 -- /usr/bin/php-cgi"
```

### 2.3 Unit-файл spawn-fcgi: `/etc/systemd/system/spawn-fcgi.service`

```ini
[Unit]
Description=Spawn-fcgi startup service by Otus
After=network.target

[Service]
Type=simple
PIDFile=/var/run/spawn-fcgi.pid
EnvironmentFile=/etc/spawn-fcgi/fcgi.conf
ExecStart=/usr/bin/spawn-fcgi -n $OPTIONS
KillMode=process

[Install]
WantedBy=multi-user.target
```

### 2.4 Запуск и проверка статуса

```console
root@new-vm:/home/faoxis# systemctl start spawn-fcgi
root@new-vm:/home/faoxis# systemctl status spawn-fcgi
● spawn-fcgi.service - Spawn-fcgi startup service by Otus
     Loaded: loaded (/etc/systemd/system/spawn-fcgi.service; disabled; preset: enabled)
     Active: active (running) since Tue 2026-02-17 22:02:15 UTC; 38s ago
   Main PID: 11361 (php-cgi)
      Tasks: 22 (limit: 4615)
     Memory: 31.4M (peak: 31.9M)
        CPU: 72ms
     CGroup: /system.slice/spawn-fcgi.service
             ├─11361 /usr/bin/php-cgi
             ├─11362 /usr/bin/php-cgi
             ├─11363 /usr/bin/php-cgi
             ├─11364 /usr/bin/php-cgi
             ├─11365 /usr/bin/php-cgi
             ├─11366 /usr/bin/php-cgi
             ├─11367 /usr/bin/php-cgi
             ├─11368 /usr/bin/php-cgi
             ├─11369 /usr/bin/php-cgi
             ├─11370 /usr/bin/php-cgi
             ├─11371 /usr/bin/php-cgi
             ├─11372 /usr/bin/php-cgi
             ├─11373 /usr/bin/php-cgi
             ├─11374 /usr/bin/php-cgi
             ├─11375 /usr/bin/php-cgi
             ├─11376 /usr/bin/php-cgi
             ├─11377 /usr/bin/php-cgi
             ├─11378 /usr/bin/php-cgi
             ├─11379 /usr/bin/php-cgi
             ├─11380 /usr/bin/php-cgi
             ├─11381 /usr/bin/php-cgi
             └─11382 /usr/bin/php-cgi
```

---

## 3) Nginx: запуск нескольких инстансов с разными конфигами через шаблонный unit

### 3.1 Шаблонный unit: `/etc/systemd/system/nginx@.service`

```ini
[Unit]
Description=A high performance web server and a reverse proxy server
Documentation=man:nginx(8)
After=network.target nss-lookup.target

[Service]
Type=forking
PIDFile=/run/nginx-%I.pid
ExecStartPre=/usr/sbin/nginx -t -c /etc/nginx/nginx-%I.conf -q -g 'daemon on; master_process on;'
ExecStart=/usr/sbin/nginx -c /etc/nginx/nginx-%I.conf -g 'daemon on; master_process on;'
ExecReload=/usr/sbin/nginx -c /etc/nginx/nginx-%I.conf -g 'daemon on; master_process on;' -s reload
ExecStop=-/sbin/start-stop-daemon --quiet --stop --retry QUIT/5 --pidfile /run/nginx-%I.pid
TimeoutStopSec=5
KillMode=mixed

[Install]
WantedBy=multi-user.target
```

### 3.2 Конфиги инстансов

`/etc/nginx/nginx-first.conf`:

```nginx
pid /run/nginx-first.pid;

events {}

http {

        server {
                listen 9001;
        }
}
```

`/etc/nginx/nginx-second.conf`:

```nginx
pid /run/nginx-second.pid;

events {}

http {

        server {
                listen 9002;
        }
}
```

### 3.3 Проверка работы

Проверка слушающих портов:

```console
root@new-vm:/home/faoxis# ss -tnulp | grep nginx
tcp   LISTEN 0      511             0.0.0.0:9002      0.0.0.0:*    users:(("nginx",pid=11571,fd=4),("nginx",pid=11570,fd=4))
tcp   LISTEN 0      511             0.0.0.0:9001      0.0.0.0:*    users:(("nginx",pid=11563,fd=4),("nginx",pid=11562,fd=4))
```

Проверка процессов:

```console
root@new-vm:/home/faoxis# ps afx | grep nginx
  11562 ?        Ss     0:00 nginx: master process /usr/sbin/nginx -c /etc/nginx/nginx-first.conf -g daemon on; master_process on;
  11563 ?        S      0:00  \_ nginx: worker process
  11570 ?        Ss     0:00 nginx: master process /usr/sbin/nginx -c /etc/nginx/nginx-second.conf -g daemon on; master_process on;
  11571 ?        S      0:00  \_ nginx: worker process
```

---

## Итог

* Реализован `watchlog` (oneshot service + timer каждые 30 секунд) с конфигом из `/etc/default`, логирование через `logger` в syslog.
* Установлен `spawn-fcgi`, создан unit-файл на основе параметров из `/etc/spawn-fcgi/fcgi.conf`, сервис запускается и находится в `active (running)`.
* Доработан запуск Nginx через шаблонный unit `nginx@.service`, одновременно работают два инстанса с разными конфигами и PID, слушают порты **9001** и **9002**.
