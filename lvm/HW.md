# Домашняя работа по теме LVM


### Список используемых технологий
1) Использовалась виртуальная машина в Google Cloud с 4 cpu и 8 RAM. К ней подключил 4 диска: 3 по 10GiB, 1 - 30GiB. Дополнительно загрузочный диск - 30GiB.
2) Используемый дистрибутив Ubuntu 24.04

### Уменьшение тома под / до 8G
* Для начал выполнил переход в режим суперпользователя командой `sudo -i`
* Создал физический раздел командой `pvcreate /dev/sdb`
  * ```shell
    root@lvm-instance1:~# pvcreate /dev/sdb
      Physical volume "/dev/sdb" successfully created.

    ```
* Создал группу томов:
  * ```shell
    root@lvm-instance1:~# vgcreate vg_root /dev/sdb
        Volume group "vg_root" successfully created
    ```
* Создал логический том:
  * ```shell
    root@lvm-instance1:~# lvcreate -n lv_root -l +100%FREE /dev/vg_root
       Logical volume "lv_root" created.
    ```
* Создал файловую систему на логическом томе
  * ```shell
    root@lvm-instance1:~# mkfs.xfs /dev/vg_root/lv_root
        meta-data=/dev/vg_root/lv_root   isize=512    agcount=4, agsize=655104 blks
        =                       sectsz=4096  attr=2, projid32bit=1
        =                       crc=1        finobt=1, sparse=1, rmapbt=1
        =                       reflink=1    bigtime=1 inobtcount=1 nrext64=0
        data     =                       bsize=4096   blocks=2620416, imaxpct=25
        =                       sunit=0      swidth=0 blks
        naming   =version 2              bsize=4096   ascii-ci=0, ftype=1
        log      =internal log           bsize=4096   blocks=16384, version=2
        =                       sectsz=4096  sunit=1 blks, lazy-count=1
        realtime =none                   extsz=4096   blocks=0, rtextents=0
        Discarding blocks...Done.
    ```
* Примонтировал файловую систему командой `mount /dev/vg_root/lv_root /mnt`
* Установил утилиту `xfsdump` командой `apt install xfsdump`
* 