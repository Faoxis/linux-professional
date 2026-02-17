#!/usr/bin/env python3
import os

HZ = os.sysconf("SC_CLK_TCK")

def parse_stat(pid: int):
    s = open(f"/proc/{pid}/stat", "r", encoding="utf-8", errors="replace").read().strip()

    i = s.find("(")
    j = s.rfind(")")
    if i == -1 or j == -1 or j <= i:
        raise ValueError("bad stat format")

    comm = s[i+1:j]
    tail = s[j+1:].split()
    state = tail[0]
    utime = int(tail[11])
    stime = int(tail[12])
    return comm, state, utime, stime

def read_cmd(pid: int):
    b = open(f"/proc/{pid}/cmdline", "rb").read()
    cmd = b.replace(b"\0", b" ").strip().decode("utf-8", errors="replace")
    return cmd

def fmt_time(sec: int):
    mm, ss = divmod(sec, 60)
    hh, mm = divmod(mm, 60)
    if hh:
        return f"{hh:02}:{mm:02}:{ss:02}"
    return f"{mm:02}:{ss:02}"

def main():
    print(f"{'PID':>5} {'STAT':<4} {'TIME':>8} COMMAND")

    pids = sorted(int(x) for x in os.listdir("/proc") if x.isdigit())
    for pid in pids:
        try:
            comm, st, ut, stt = parse_stat(pid)
            cmd = read_cmd(pid)
            if not cmd:
                cmd = f"[{comm}]"

            sec = (ut + stt) // HZ
            t = fmt_time(sec)

            print(f"{pid:5d} {st:<4} {t:>8} {cmd}")
        except Exception:
            pass

if __name__ == "__main__":
    main()
