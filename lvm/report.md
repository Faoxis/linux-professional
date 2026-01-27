# Отчёт по ДЗ «Файловые системы и LVM» (Ubuntu 24.04, MWS облако)

## Контекст
- Исходно система была установлена **без LVM**: root (`/`) на `/dev/sda1` (ext4), отдельные `/boot` и `/boot/efi`.
- Для выполнения ДЗ были добавлены отдельные блочные устройства и на них создан LVM.

## 1) Исходное состояние дисков
Команда проверки разметки и подключенных дисков:

```bash
lsblk
```

```text
NAME    MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
sda       8:0    0   10G  0 disk
├─sda1    8:1    0    9G  0 part /
├─sda14   8:14   0    4M  0 part
├─sda15   8:15   0  106M  0 part /boot/efi
└─sda16 259:0    0  913M  0 part /boot
sdb       8:16   0   50K  1 disk
sdc       8:32   0  100G  0 disk
sdd       8:48   0  100G  0 disk
sde       8:64   0  100G  0 disk
sdf       8:80   0  100G  0 disk
```

## 2) Создание LVM (PV → VG)

Создаём physical volumes и volume group (далее использовался VG `otus_hw`).

Проверка PV:

```bash
pvs
```

```text
PV         VG      Fmt  Attr PSize    PFree
  /dev/sdc   otus_hw lvm2 a--  <100.00g <100.00g
  /dev/sdd   otus_hw lvm2 a--  <100.00g <100.00g
  /dev/sde   otus_hw lvm2 a--  <100.00g <100.00g
  /dev/sdf   otus_hw lvm2 a--  <100.00g <100.00g
```

Проверка VG:

```bash
vgs
```

```text
VG      #PV #LV #SN Attr   VSize   VFree
  otus_hw   4   0   0 wz--n- 399.98g 399.98g
```

## 3) Переезд системы на LVM: временный root на `lv_root` (100G)

### 3.1 Создать LV под root

```bash
lvcreate -L 100G -n lv_root otus_hw
```

```text
Logical volume "lv_root" created.
```

Проверка, что LV появился:

```bash
lvs -o lv_name,vg_name,lv_size,lv_attr
```

```text
lvs -o lv_name,vg_name,lv_size,lv_attr



  LV      VG      LSize   Attr
  lv_root otus_hw 100.00g -wi-a-----
```

### 3.2 Создать файловую систему и смонтировать в `/mnt`

`/mnt` — это обычный каталог, который традиционно используют как *временную точку монтирования*. Мы монтируем туда новый LV, чтобы скопировать в него содержимое текущего `/`.

```bash
mkfs.ext4 -L rootfs /dev/mapper/otus_hw-lv_root
```

```text
mke2fs 1.47.0 (5-Feb-2023)
Discarding device blocks: done
Creating filesystem with 26214400 4k blocks and 6553600 inodes
Filesystem UUID: 6974c84e-37ab-4545-8da0-a1a06d71f2e8
Superblock backups stored on blocks:
	32768, 98304, 163840, 229376, 294912, 819200, 884736, 1605632, 2654208,
	4096000, 7962624, 11239424, 20480000, 23887872

Allocating group tables: done
Writing inode tables: done
Creating journal (131072 blocks): done
Writing superblocks and filesystem accounting information: done
```

### 3.3 Скопировать текущий root на новый LV

Копирование делалось `rsync` с сохранением прав/ACL/xattrs и без перехода на другие FS (`-x`). Спецкаталоги `/proc`, `/sys`, `/dev`, `/run` не копируются — почему, объяснение ниже.

```bash
rsync -aHAXx --numeric-ids --info=progress2 \
  --exclude="/mnt/*" \
  --exclude="/proc/*" --exclude="/sys/*" --exclude="/dev/*" --exclude="/run/*" \
  --exclude="/tmp/*" \
  / /mnt/
```

```text
  1,971,579,495  99%   22.09MB/s    0:01:25 (xfr#69678, to-chk=0/85281)
```


### 3.4 Подготовить chroot: bind-mount и обновить загрузку

Чтобы внутри `chroot /mnt` работали утилиты, которым нужен доступ к `/dev`, `/proc`, `/sys`, `/run`, мы **пробиндим** (bind-mount) эти каталоги в новый корень:

```bash
mount --bind /dev  /mnt/dev
mount --bind /run  /mnt/run
mount --bind /proc /mnt/proc
mount --bind /sys  /mnt/sys
```

Далее:
- `chroot /mnt/`
- пересобрать конфиг GRUB
- обновить initramfs

Вывод ключевых команд:

```bash
grub-mkconfig -o /boot/grub/grub.cfg
```

```text
grub-mkconfig -o /boot/grub/grub.cfg

Sourcing file `/etc/default/grub'
Sourcing file `/etc/default/grub.d/50-cloudimg-settings.cfg'
Generating grub configuration file ...
Found linux image: /boot/vmlinuz-6.8.0-60-generic
Found initrd image: /boot/initrd.img-6.8.0-60-generic
Warning: os-prober will not be executed to detect other bootable partitions.
Systems on them will not be added to the GRUB boot configuration.
Check GRUB_DISABLE_OS_PROBER documentation entry.
Adding boot menu entry for UEFI Firmware Settings ...
done
```
```bash
update-initramfs -u
```

```text
update-initramfs: Generating /boot/initrd.img-6.8.0-60-generic
```

### 3.5 Перезагрузка и проверка, что `/` теперь на LVM

Проверка по `lsblk`, `df`, `lvs`:

```bash
lsblk
```

```text
NAME              MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
sda                 8:0    0   10G  0 disk
├─sda1              8:1    0    9G  0 part
├─sda14             8:14   0    4M  0 part
├─sda15             8:15   0  106M  0 part /boot/efi
└─sda16           259:0    0  913M  0 part /boot
sdb                 8:16   0  100G  0 disk
└─otus_hw-lv_root 252:0    0  100G  0 lvm  /
sdc                 8:32   0  100G  0 disk
└─otus_hw-lv_root 252:0    0  100G  0 lvm  /
sdd                 8:48   0  100G  0 disk
sde                 8:64   0  100G  0 disk
sdf                 8:80   0   50K  1 disk
```
```bash
df -hT
```

```text
Filesystem                  Type      Size  Used Avail Use% Mounted on
tmpfs                       tmpfs     1.6G  1.2M  1.6G   1% /run
efivarfs                    efivarfs  256K   23K  229K   9% /sys/firmware/efi/efivars
/dev/mapper/otus_hw-lv_root ext4       98G  2.1G   91G   3% /
tmpfs                       tmpfs     7.8G     0  7.8G   0% /dev/shm
tmpfs                       tmpfs     5.0M     0  5.0M   0% /run/lock
/dev/sda16                  ext4      881M   62M  758M   8% /boot
/dev/sda15                  vfat      105M  6.2M   99M   6% /boot/efi
tmpfs                       tmpfs     1.6G   12K  1.6G   1% /run/user/1000
```
```bash
lvs
```

```text
LV      VG      Attr       LSize   Pool Origin Data%  Meta%  Move Log Cpy%Sync Convert
  lv_root otus_hw -wi-ao---- 100.00g
root@lvm-y067np:~#

root@lvm-y067np:~#

root@lvm-y067np:~#
```

## 4) Уменьшение root до 8G: переезд на `real_root` (8G)

Так как root уже находится на LVM, «уменьшение» выполняется безопасным способом: создаём новый LV нужного размера, копируем туда систему, обновляем загрузку, перезагружаемся.


### 4.1 Создать `real_root` на 8G

```bash
lvcreate -n real_root -L 8G otus_hw
```

```text
Logical volume "real_root" created.
```

### 4.2 Создать ФС, смонтировать и скопировать `/`

```bash
mkfs.ext4 /dev/otus_hw/real_root
```

```text
mke2fs 1.47.0 (5-Feb-2023)
Discarding device blocks: done
Creating filesystem with 2097152 4k blocks and 524288 inodes
Filesystem UUID: 4f30aa01-a6b8-415d-9031-09b125badd06
Superblock backups stored on blocks:
	32768, 98304, 163840, 229376, 294912, 819200, 884736, 1605632

Allocating group tables: done
Writing inode tables: done
Creating journal (16384 blocks): done
Writing superblocks and filesystem accounting information: done
```

Копирование:

```bash
rsync -avxHAX --progress / /mnt/
```

```text

sent 2,009,076,384 bytes  received 1,394,442 bytes  53,612,555.36 bytes/sec
total size is 2,007,544,982  speedup is 1.00
```

### 4.3 Обновить GRUB/initramfs из chroot и перезагрузиться

```bash
grub-mkconfig -o /boot/grub/grub.cfg
```

```text
grub-mkconfig -o /boot/grub/grub.cfg

Sourcing file `/etc/default/grub'
Sourcing file `/etc/default/grub.d/50-cloudimg-settings.cfg'
Generating grub configuration file ...
Found linux image: /boot/vmlinuz-6.8.0-60-generic
Found initrd image: /boot/initrd.img-6.8.0-60-generic
Warning: os-prober will not be executed to detect other bootable partitions.
Systems on them will not be added to the GRUB boot configuration.
Check GRUB_DISABLE_OS_PROBER documentation entry.
Adding boot menu entry for UEFI Firmware Settings ...
done
```
```bash
update-initramfs -u
```

```text
update-initramfs -u



update-initramfs: Generating /boot/initrd.img-6.8.0-60-generic
```

### 4.4 Проверка после перезагрузки: `/` на `real_root` (8G)

```bash
lsblk
```

```text
NAME                MAJ:MIN RM  SIZE RO TYPE MOUNTPOINTS
sda                   8:0    0   10G  0 disk
├─sda1                8:1    0    9G  0 part
├─sda14               8:14   0    4M  0 part
├─sda15               8:15   0  106M  0 part /boot/efi
└─sda16             259:0    0  913M  0 part /boot
sdb                   8:16   0  100G  0 disk
└─otus_hw-lv_root   252:0    0  100G  0 lvm
sdc                   8:32   0  100G  0 disk
├─otus_hw-lv_root   252:0    0  100G  0 lvm
└─otus_hw-real_root 252:1    0    8G  0 lvm  /
sdd                   8:48   0  100G  0 disk
sde                   8:64   0  100G  0 disk
sdf                   8:80   0   50K  1 disk
```

### 4.5 Удалить старый большой `lv_root`, освободить место

```bash
lvremove /deotus_hw/lv_root
```

```text
Do you really want to remove and DISCARD active logical volume otus_hw/lv_root? [y/n]: y
  Logical volume "lv_root" successfully removed.
```

Проверка LVs:

```bash
lvs
```

```text
LV        VG      Attr       LSize Pool Origin Data%  Meta%  Move Log Cpy%Sync Convert
  real_root otus_hw -wi-ao---- 8.00g
```

## 5) `/var` отдельным томом в mirror (LVM RAID1)

Создаём LV `lv_var` с зеркалированием (`-m1` = 1 mirror → 2 копии данных на разных PV).

```bash
lvcreate -L 950M -m1 -n lv_var otus_hw
```

```text
lvcreate -L 950M -m1 -n lv_var otus_hw

  Rounding up size to full physical extent 952.00 MiB
  Logical volume "lv_var" created.
```

Создание ФС на `/var` (ext4):

```bash
mkfs.ext4 /dev/otus_hw
```

```text
mke2fs 1.47.0 (5-Feb-2023)
Discarding device blocks: done
Creating filesystem with 243712 4k blocks and 60928 inodes
Filesystem UUID: 2d721f2c-74a7-48a2-8b89-c4aabe185412
Superblock backups stored on blocks:
	32768, 98304, 163840, 229376

Allocating group tables: done
Writing inode tables: done
Creating journal (4096 blocks): done
Writing superblocks and filesystem accounting information: done
```

Дальше по методике выглядит так:

```bash
mount /dev/otus_hw/lv_var /mnt
cp -aR /var/* /mnt/
mkdir -p /tmp/oldvar && mv /var/* /tmp/oldvar
umount /mnt
mount /dev/otus_hw/lv_var /var
```

Добавление в `/etc/fstab`:

```bash
echo "`blkid | grep var: | awk '{print $2}'` /var ext4 defaults 0 0" >> /etc/fstab
systemctl daemon-reload
```

## 6) `/home` отдельным томом + снапшоты LVM

### 6.1 Создать LV под `/home` и ФС

```bash
lvcreate -n lv_home -L 50G otus_hw
```

```bash
mkfs.ext4 /dev/otus_hw/lv_home
```
```text
mke2fs 1.47.0 (5-Feb-2023)
Discarding device blocks: done
Creating filesystem with 13107200 4k blocks and 3276800 inodes
Filesystem UUID: f85fdd94-8057-4206-8fa0-fcda74e89e48
Superblock backups stored on blocks:
	32768, 98304, 163840, 229376, 294912, 819200, 884736, 1605632, 2654208,
	4096000, 7962624, 11239424

Allocating group tables: done
Writing inode tables: done
Creating journal (65536 blocks): done
Writing superblocks and filesystem accounting information: done
```

### 6.2 Перенос `/home` и fstab

```bash
mount /dev/otus_hw/lv_home /mnt
cp -aR /home/* /mnt/
rm -rf /home/*
umount /mnt
mount /dev/otus_hw/lv_home /home

echo "`blkid | grep Home | awk '{print $2}'`  /home xfs defaults 0 0" >> /etc/fstab
systemctl daemon-reload
```

### 6.3 Снапшот → удаление части файлов → восстановление

Проверка содержимого папки home:

```bash
ls /home/
```

```text
faoxis  lost+found
```


Создание снимка:

```bash
lvcreate -L 1G -s -n home_snap /dev/otus_hw/homlv_home
```

```text
lost+found
```

Удаляем часть файлов:

```bash
rm -rf /home/faoxis
```

Проверим, что содержимое удалилось:

```bash
ls /home/
```

```text
Logical volume "home_snap" created.
```


Откат:

```bash
umount /home
lvconvert --merge /dev/otus_hw/home_snap
mount /dev/otus_hw/lv_home /home
```

```text
mount: (hint) your fstab has been modified, but systemd still uses
       the old version; use 'systemctl daemon-reload' to reload.
```

Проверка, что файлы восстановились:

```bash
ls /home/
```

```text
faoxis  lost+found
```
