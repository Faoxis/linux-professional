#### Выполнение домашнего задания по теме "Vagrant"

1) Создание [`Vagrantfile`](./Vagrantfile) c образом `bento/ubuntu-24.04`
2) Переход в виртуальную машину командой `vagrant ssh`
3) Проверка текущей версии ядра `uname -r`: `6.8.0-31-generic`
4) Обновление репозитория: `sudo apt update && sudo apt upgrade -y`
5) Установка доступного обновления ядра: `sudo apt install --install-recommends linux-generic`
6) Обновление конфигурации загрузчика `grub`: `sudo update-grub` 
7) Проверка, что новое ядро загрузилось и стало основным командой `ls /boot/vmlinuz*`:
```
/boot/vmlinuz                   /boot/vmlinuz-6.8.0-48-generic
/boot/vmlinuz-6.8.0-31-generic  /boot/vmlinuz.old
```
8) Перезапуск виртуальной машины: `sudo reboot`
9) Подключимся к ВМ командой `vagrant ssh` и убедимся, что работа ведется на обновленном ядре командой `uname -r`: `6.8.0-48-generic`

Готово