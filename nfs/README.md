# Работа с NFS 

## Настройка сервера
Подключаемся к серверу по ssh и запускает [скрипт](server.sh) настройки nfs сервера.
Вывод:
```shell
[2/6] Checking listening ports (2049, 111) ...
udp   UNCONN 0      0              0.0.0.0:111        0.0.0.0:*    users:(("rpcbind",pid=2012,fd=5),("systemd",pid=1,fd=133))
udp   UNCONN 0      0                 [::]:111           [::]:*    users:(("rpcbind",pid=2012,fd=7),("systemd",pid=1,fd=140))
tcp   LISTEN 0      64             0.0.0.0:2049       0.0.0.0:*
tcp   LISTEN 0      4096           0.0.0.0:111        0.0.0.0:*    users:(("rpcbind",pid=2012,fd=4),("systemd",pid=1,fd=128))
tcp   LISTEN 0      64                [::]:2049          [::]:*
tcp   LISTEN 0      4096              [::]:111           [::]:*    users:(("rpcbind",pid=2012,fd=6),("systemd",pid=1,fd=137))
[3/6] Creating export directories...
[4/6] Setting permissions...
[5/6] Writing /etc/exports ...
[6/6] Applying exports and showing result...
exportfs: /etc/exports [1]: Neither 'subtree_check' or 'no_subtree_check' specified for export "10.0.0.9:/srv/share".
  Assuming default behaviour ('no_subtree_check').
  NOTE: this default has changed since nfs-utils version 1.0.x

/srv/share  10.0.0.9(sync,wdelay,hide,no_subtree_check,sec=sys,rw,secure,root_squash,no_all_squash)
Done.
```

## Настраиваем клиент
Подключаемся к серверу по ssh и запускает [скрипт](client.sh) настройки nfs клиента.
Вывод:
```shell
[2/6] Ensuring mount point exists: /mnt
[3/6] Adding /etc/fstab entry (idempotent)...
[4/6] Reloading systemd units...
[5/6] Restarting remote-fs.target...
[6/6] Triggering automount (first access to /mnt)...
Mount status (expect autofs + nfs lines if success):
systemd-1 on /mnt type autofs (rw,relatime,fd=67,pgrp=1,timeout=0,minproto=5,maxproto=5,direct,pipe_ino=18435)
10.0.0.8:/srv/share on /mnt type nfs (rw,relatime,vers=3,rsize=1048576,wsize=1048576,namlen=255,hard,proto=tcp,timeo=600,retrans=2,sec=sys,mountaddr=10.0.0.8,mountvers=3,mountport=49157,mountproto=udp,local_lock=none,addr=10.0.0.8,_netdev)
Done.
```


## Проверяем работоспособность скриптов

### Проверка видимости файлов
- Заходим на сервер и выполняем команду `touch /srv/share/upload/server_check`
- Заходим на клиент и выполняем команду `ll /mnt/upload/`
```shell
total 8
drwxrwxrwx 2 nobody nogroup 4096 Feb  5 07:53 ./
drwxr-xr-x 3 nobody nogroup 4096 Feb  5 07:43 ../
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 07:53 server_check
```
- Выполним команду `touch /mnt/upload/client_check`
- Проверим наличие файлов на сервере `ll /srv/share/upload/`
```shell
total 8
drwxrwxrwx 2 nobody nogroup 4096 Feb  5 07:54 ./
drwxr-xr-x 3 nobody nogroup 4096 Feb  5 07:43 ../
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 07:54 client_check
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 07:53 server_check
```
- Делаем вывод, что клиент-серверное взаимодействие работает

### Устойчивость к перезапуску клиента
- Заходим на клиент и выполняем команду `sudo reboot`
- Снова подключаемся к клиенту и выполняем команду `ll /mnt/upload/`
```shell
total 8
drwxrwxrwx 2 nobody nogroup 4096 Feb  5 07:54 ./
drwxr-xr-x 3 nobody nogroup 4096 Feb  5 07:43 ../
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 07:54 client_check
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 07:53 server_check
```
- Выполним команду `showmount -a 10.0.0.8`
``shell
All mount points on 10.0.0.8:
10.0.0.9:/srv/share
``
- Проверим точку монтирования `mount | grep mnt` 
```shell
systemd-1 on /mnt type autofs (rw,relatime,fd=57,pgrp=1,timeout=0,minproto=5,maxproto=5,direct,pipe_ino=5832)
10.0.0.8:/srv/share on /mnt type nfs (rw,relatime,vers=3,rsize=1048576,wsize=1048576,namlen=255,hard,proto=tcp,timeo=600,retrans=2,sec=sys,mountaddr=10.0.0.8,mountvers=3,mountport=49157,mountproto=udp,local_lock=none,addr=10.0.0.8,_netdev)
```
- Делаем вывод, что все работает

### Устойчивость к перезапуску сервера
- Заходим на сервер и выполнил команду `sudo reboot`
- После перезапуска заходим на сервер и выполним команду `ll /srv/share/upload/`
```shell
total 8
drwxrwxrwx 2 nobody nogroup 4096 Feb  5 07:54 ./
drwxr-xr-x 3 nobody nogroup 4096 Feb  5 07:43 ../
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 07:54 client_check
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 07:53 server_check
```
- Проверим экспорты `exportfs -s`
```shell
/srv/share  10.0.0.9(sync,wdelay,hide,no_subtree_check,sec=sys,rw,secure,root_squash,no_all_squash)
```
- Проверим работу RPC командой `sudo showmount -a 10.0.0.8`
```shell
All mount points on 10.0.0.8:
10.0.0.9:/srv/share
```
- Создадим файл `final_check` на клиент и выполним на сервер команду `ll /srv/share/upload/`
```shell
total 8
drwxrwxrwx 2 nobody nogroup 4096 Feb  5 08:07 ./
drwxr-xr-x 3 nobody nogroup 4096 Feb  5 07:43 ../
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 07:54 client_check
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 08:07 final_check
-rw-rw-r-- 1 faoxis faoxis     0 Feb  5 07:53 server_check
```


