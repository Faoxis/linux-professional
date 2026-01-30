# Отчёт по ДЗ: Практические навыки работы с ZFS

**Цель:** научиться устанавливать и использовать ZFS, настраивать пулы, сравнивать сжатие, импортировать pool, работать со снапшотами/receive.

---

## Окружение и подготовка

### Проверка дисков

Домашняя работа полностью проделана в облаке MWS Cloud. Минимальный размер диска в облаке - 1G. Делал исходя из этого.

Команда:

```bash
lsblk
```

Вывод:

```text
NAME    MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
sda       8:0    0   10G  0 disk
├─sda1    8:1    0    9G  0 part /
├─sda14   8:14   0    4M  0 part
├─sda15   8:15   0  106M  0 part /boot/efi
└─sda16 259:0    0  913M  0 part /boot
sdb       8:16   0   50K  1 disk
sdc       8:32   0    1G  0 disk
sdd       8:48   0    1G  0 disk
sde       8:64   0    1G  0 disk
sdf       8:80   0    1G  0 disk
sdg       8:96   0    1G  0 disk
sdh       8:112  0    1G  0 disk
sdi       8:128  0    1G  0 disk
sdj       8:144  0    1G  0 disk
```

---

## 1) Определить алгоритм с наилучшим сжатием

### 1.1 Создание пулов (mirror)

Команды:

```bash
zpool create otus1 mirror sdc sdd
zpool create otus2 mirror sde sdf
zpool create otus3 mirror sdg sdh
zpool create otus4 mirror sdi sdj
```

Проверка:

```bash
zpool list
```

Вывод:

```text
NAME    SIZE  ALLOC   FREE  CKPOINT  EXPANDSZ   FRAG    CAP  DEDUP    HEALTH  ALTROOT
otus1   960M   420K   960M        -         -     0%     0%  1.00x    ONLINE  -
otus2   960M   408K   960M        -         -     0%     0%  1.00x    ONLINE  -
otus3   960M   432K   960M        -         -     0%     0%  1.00x    ONLINE  -
otus4   960M   444K   960M        -         -     0%     0%  1.00x    ONLINE  -
```

Проверка состояния:

```bash
zpool status
```

Вывод (фрагмент):

```text
pool: otus1
  mirror-0
    sdc  ONLINE
    sdd  ONLINE

pool: otus2
  mirror-0
    sde  ONLINE
    sdf  ONLINE

pool: otus3
  mirror-0
    sdg  ONLINE
    sdh  ONLINE

pool: otus4
  mirror-0
    sdi  ONLINE
    sdj  ONLINE

errors: No known data errors
```

---

### 1.2 Назначение разных алгоритмов compression

Команды:

```bash
zfs set compression=lzjb otus1
zfs set compression=lz4  otus2
zfs set compression=gzip-9 otus3
zfs set compression=zle  otus4
```

Проверка:

```bash
zfs get all | grep compression
```

Вывод:

```text
otus1  compression           lzjb                   local
otus2  compression           lz4                    local
otus3  compression           gzip-9                 local
otus4  compression           zle                    local
```

---

### 1.3 Тест сжатия на одном и том же файле

Скачивание одинакового файла в каждый пул:

```bash
for i in {1..4}; do wget -P /otus$i https://gutenberg.org/cache/epub/2600/pg2600.converter.log; done
```

Проверка наличия и «видимого» размера (как хранится в FS):

```bash
ls -l /otus*
```

Вывод:

```text
/otus1:
total 22557
-rw-r--r-- 1 root root 41204725 Jan  2 08:31 pg2600.converter.log

/otus2:
total 18509
-rw-r--r-- 1 root root 41204725 Jan  2 08:31 pg2600.converter.log

/otus3:
total 11477
-rw-r--r-- 1 root root 41204725 Jan  2 08:31 pg2600.converter.log

/otus4:
total 40277
-rw-r--r-- 1 root root 41204725 Jan  2 08:31 pg2600.converter.log
```

Сравнение реального потребления места ZFS:

```bash
zfs list
```

Вывод:

```text
NAME    USED  AVAIL  REFER  MOUNTPOINT
otus1  22.6M   809M  22.1M  /otus1
otus2  18.6M   813M  18.2M  /otus2
otus3  11.7M   820M  11.3M  /otus3
otus4  39.8M   792M  39.4M  /otus4
```

Коэффициент сжатия:

```bash
zfs get all | grep compressratio | grep -v ref
```

Вывод:

```text
otus1  compressratio         1.78x                  -
otus2  compressratio         2.16x                  -
otus3  compressratio         3.47x                  -
otus4  compressratio         1.00x                  -
```

**Вывод по заданию 1:** наиболее эффективный алгоритм сжатия — **gzip-9** (pool `otus3`), `compressratio=3.47x`, минимальное потребление места (`USED=11.7M`).

---

## 2) Определение настроек пула через zpool import + zfs/zpool get

### 2.1 Скачивание и распаковка экспортированного пула

Скачать архив:

```bash
wget -O archive.tar.gz --no-check-certificate 'https://drive.usercontent.google.com/download?id=1MvrcEp-WgAQe57aDEzxSRalPAwbNN1Bb&export=download'
```

Распаковать:

```bash
tar -xzvf archive.tar.gz
```

Вывод:

```text
zpoolexport/
zpoolexport/filea
zpoolexport/fileb
```

---

### 2.2 Поиск пула для импорта и импорт

Проверка наличия пула в каталоге:

```bash
zpool import -d zpoolexport/
```

Вывод:

```text
pool: otus
  id: 6554193320433390805
state: ONLINE
config:

	otus                         ONLINE
	  mirror-0                   ONLINE
	    /root/zpoolexport/filea  ONLINE
	    /root/zpoolexport/fileb  ONLINE
```

Импорт:

```bash
zpool import -d zpoolexport/ otus
```

Проверка статуса:

```bash
zpool status
```

Вывод (фрагмент):

```text
pool: otus
state: ONLINE
config:
	NAME                         STATE     READ WRITE CKSUM
	otus                         ONLINE       0     0     0
	  mirror-0                   ONLINE       0     0     0
	    /root/zpoolexport/filea  ONLINE       0     0     0
	    /root/zpoolexport/fileb  ONLINE       0     0     0
errors: No known data errors
```

---

### 2.3 Определение параметров пула

Общие параметры пула:

```bash
zpool get all otus
```

Вывод (фрагмент ключевого):

```text
otus  size        480M
otus  free        478M
otus  allocated   2.09M
otus  health      ONLINE
otus  guid        6554193320433390805
```

Требуемые параметры из задания:

**Размер доступного хранилища:**

```bash
zfs get available otus
```

Вывод:

```text
NAME  PROPERTY   VALUE  SOURCE
otus  available  350M   -
```

**Тип пула (RAID/топология vdev):**

* По `zpool status` видно `mirror-0` → тип: **mirror**

**Проверка readonly:**

```bash
zfs get readonly otus
```

Вывод:

```text
NAME  PROPERTY  VALUE   SOURCE
otus  readonly  off     default
```

**recordsize:**

```bash
zfs get recordsize otus
```

Вывод:

```text
NAME  PROPERTY    VALUE    SOURCE
otus  recordsize  128K     local
```

**compression:**

```bash
zfs get compression otus
```

Вывод:

```text
NAME  PROPERTY     VALUE  SOURCE
otus  compression  zle    local
```

**checksum:**

```bash
zfs get checksum otus
```

Вывод:

```text
NAME  PROPERTY  VALUE   SOURCE
otus  checksum  sha256  local
```

**Вывод по заданию 2 (кратко):**

* Pool: `otus`
* Topology: **mirror**
* Size: `480M`
* Available: `350M`
* recordsize: `128K`
* compression: `zle`
* checksum: `sha256`
* readonly: `off`

---

## 3) Работа со снапшотами и восстановлением (zfs receive) + поиск secret_message

### 3.1 Скачивание файла для receive

Команда:

```bash
wget -O otus_task2.file --no-check-certificate "https://drive.usercontent.google.com/download?id=1wgxjih8YZ-cqLqaZVa0lA3h3Y029c3oI&export=download"
```

> В логе запускался `wget` с записью в `wget-log`.

### 3.2 Восстановление датасета из потока (receive)

Команда:

```bash
zfs receive otus/test@today < otus_task2.file
```

### 3.3 Поиск секретного сообщения

Найти файл `secret_message`:

```bash
find /otus/test -name "secret_message"
```

Вывод:

```text
/otus/test/task1/file_mess/secret_message
```

Посмотреть содержимое:

```bash
cat /otus/test/task1/file_mess/secret_message
```

Вывод:

```text
https://otus.ru/lessons/linux-hl/
```

**Вывод по заданию 3:** датасет восстановлен через `zfs receive`, файл `secret_message` найден и прочитан.

---

## Итог

Все пункты ДЗ выполнены:

1. Сравнение алгоритмов сжатия: лучший **gzip-9** (`compressratio=3.47x`, min USED).
2. Импорт пула `otus` через `zpool import`, определены параметры (mirror, recordsize=128K, compression=zle, checksum=sha256, available=350M).
3. Восстановление через `zfs receive`, найдено `secret_message` и извлечено значение.
