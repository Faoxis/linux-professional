# Домашнее задание: PXE-сервер с автоустановкой Ubuntu 24

## Задание

- Настроить загрузку по сети дистрибутива Ubuntu 24.
- Установка должна проходить из HTTP-репозитория.
- Настроить автоматическую установку с помощью файла `user-data`.

## Стенд

Две машины на bento/ubuntu-22.04:

- `pxeserver` - раздаёт DHCP, TFTP и HTTP. Два интерфейса: NAT и внутренняя сеть `pxenet`,
  в которой у него 10.0.0.20.
- `pxeclient` - чистая машина без ОС, первым в boot order стоит сеть.

Всё, что делает сервер, собрано в `ansible/provision.yml`, он запускается провижинером из `Vagrantfile`:

```
ansible/
├── provision.yml          установка и настройка dnsmasq + apache
└── files/
    ├── pxe.conf           конфиг dnsmasq (DHCP + TFTP)
    ├── ks-server.conf     vhost apache на /srv/ks и /srv/images
    ├── pxelinux-default   pxelinux.cfg/default с параметрами автоустановки
    └── user-data          cloud-init для subiquity
```

## Перед запуском

VirtualBox должен уметь работать с 10.0.0.0/8, иначе vagrant не поднимет приватную сеть.
В `/etc/vbox/networks.conf`:

```
* 10.0.0.0/8
* 192.168.0.0/16
```

Если хост на AMD и в ядре загружен `kvm_amd`, VirtualBox не сможет забрать AMD-V. При этом
`vagrant up` пишет не про виртуализацию, а что-то вроде

```
VBoxManage: error: The machine 'DHCP_PXE_pxeserver_...' is already locked by a session
```

Настоящая причина видна только в логе самой машины (`~/VirtualBox VMs/<vm>/Logs/VBox.log`):

```
VMSetError: VirtualBox can't enable the AMD-V extension. Please disable the KVM kernel extension (VERR_SVM_IN_USE)
```

Лечится выгрузкой модуля:

```bash
sudo modprobe -r kvm_amd
```

## Запуск

```bash
vagrant up
```

Сначала поднимается `pxeserver` и на нём отрабатывает плейбук: ставит dnsmasq и apache2, гасит ufw,
качает netboot-архив в `/srv/tftp` и iso-образ (3.2 ГБ) в `/srv/images`, раскладывает `user-data` и
конфиги. Вывод:

```
RUNNING HANDLER [Перезапустить dnsmasq] ****************************************
changed: [pxeserver]
RUNNING HANDLER [Перезапустить apache2] ****************************************
changed: [pxeserver]

PLAY RECAP *********************************************************************
pxeserver                  : ok=16   changed=13   unreachable=0    failed=0    skipped=0
```

Дальше стартует `pxeclient` и уходит в PXE. На нём `vagrant up` завершится ошибкой: vagrant ждёт
ssh, а на машине в это время идёт установка ОС. Ошибку игнорируем и просто ждём.

Когда установка закончится, `pxeclient` выключится сам (`shutdown: poweroff` в `user-data`).

## Загрузка установленной системы

Просто так `vagrant up pxeclient` тут не поможет: vagrant на каждом запуске заново применяет
`vb.customize` из `Vagrantfile`, а там первым стоит `--boot1 net`. Машина снова уйдёт в PXE и
переустановится по кругу. Поэтому boot order переставляем руками, через GUI VirtualBox или так:

```bash
VBoxManage modifyvm <имя_вм> --boot1 disk --boot2 none
VBoxManage startvm <имя_вм> --type headless
```

## Проверка

Сервисы на сервере:

```shell
vagrant@pxeserver:~$ systemctl is-active dnsmasq apache2
active
active
vagrant@pxeserver:~$ sudo ss -lunp | grep -E ':67|:69'
UNCONN 0  0     0.0.0.0%eth1:67   0.0.0.0:*  users:(("dnsmasq",pid=788,fd=4))
UNCONN 0  0        10.0.0.20:69   0.0.0.0:*  users:(("dnsmasq",pid=788,fd=8))
```

Отдача по HTTP (206 на iso, потому что запрашивал только первый килобайт):

```shell
vagrant@pxeserver:~$ curl -so /dev/null -w '%{http_code}\n' http://10.0.0.20/srv/ks/user-data
200
vagrant@pxeserver:~$ curl -so /dev/null -w '%{http_code}\n' -r 0-1024 http://10.0.0.20/srv/images/noble-live-server-amd64.iso
206
```

Как клиент грузится по сети, видно в логе dnsmasq. Сначала он получает адрес и тянет загрузчик
по TFTP, потом уже поднявшийся installer заново берёт адрес:

```shell
vagrant@pxeserver:~$ sudo journalctl -u dnsmasq --no-pager | tail
dnsmasq-tftp[788]: sent /srv/tftp/amd64/pxelinux.cfg/default to 10.0.0.117
dnsmasq-tftp[788]: sent /srv/tftp/amd64/linux to 10.0.0.117
dnsmasq-tftp[788]: sent /srv/tftp/amd64/initrd to 10.0.0.117
dnsmasq-dhcp[788]: DHCPDISCOVER(eth1) 08:00:27:16:dd:ef
dnsmasq-dhcp[788]: DHCPOFFER(eth1) 10.0.0.117 08:00:27:16:dd:ef
dnsmasq-dhcp[788]: DHCPREQUEST(eth1) 10.0.0.117 08:00:27:16:dd:ef
dnsmasq-dhcp[788]: DHCPACK(eth1) 10.0.0.117 08:00:27:16:dd:ef ubuntu-server
```

На экране клиента в это время идёт subiquity: применяет `user-data` и передаёт управление curtin.

```
finish: subiquity/Identity/apply_autoinstall_config:
finish: subiquity/SSH/apply_autoinstall_config:
finish: subiquity/Shutdown/apply_autoinstall_config:
start: subiquity/Install/install/curtin_install: installing system
```

После установки и перезагрузки с диска система представляется тем именем, которое задано
в `user-data`:

```shell
dnsmasq-dhcp[788]: DHCPACK(eth1) 10.0.0.119 08:00:27:16:dd:ef linux
```

Заходим на неё с сервера под пользователем `otus` (пароль `123`, в `user-data` лежит его SHA512):

```shell
vagrant@pxeserver:~$ ssh otus@10.0.0.119
otus@linux:~$ hostname
linux
otus@linux:~$ lsb_release -ds
Ubuntu 24.04.4 LTS
otus@linux:~$ uname -r
6.8.0-137-generic
otus@linux:~$ ip -br addr | grep -v lo
enp0s3    UP    10.0.0.119/24 metric 100 fe80::a00:27ff:fe16:ddef/64
enp0s8    UP    10.0.3.15/24 metric 100 fe80::a00:27ff:fe65:bfa5/64
```

Ubuntu 24.04 поставилась по сети из HTTP-образа полностью без участия рук, задание выполнено.


