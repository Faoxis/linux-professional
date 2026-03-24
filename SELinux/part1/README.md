# Домашнее задание: SELinux

## Запуск nginx на нестандартном порту

Нужно запустить nginx на порту 4881 тремя разными способами при включённом SELinux (Enforcing).

Развернул ВМ в облаке и настроил хост с Centos/stream9:

```bash
ansible-playbook playbook.yml
```

### Способ 1 — setsebool

```bash
ansible-playbook method1.yml
```

[Более подробное описание процесса](./method1.yml).


### Способ 2 — semanage port

Добавляем наш порт в список разрешённых для http:

```bash
ansible-playbook method2.yml
```

[Более подробное описание процесса](./method2.yml).

### Способ 3 — модуль SELinux через audit2allow

```bash
ansible-playbook method3.yml
```

[Более подробное описание процесса](./method3.yml).
