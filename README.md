# Golan

```
 _______
< dns ?! >
 -------
  \
¯\_(ツ)_/¯
```


Self-hosted wildcard DNS for local networks — like nip.io or traefik.me but on your own infra. Pairs with dnsmasq to give every device on your lab a clean, resolvable hostname automatically.

```
$ dig 192.168.0.137.local.lan
192.168.0.137

$ dig -x 192.168.0.137
192.168.0.137.local.lan.
```

## How it works

- Any prefix labels are ignored — `app1.192.168.0.137.local.lan` and `app2.192.168.0.137.local.lan` both resolve to the same IP.
- Unknown return `127.0.0.1`.
- Reverse DNS (`-x`) is supported.

## Build

```bash
make
make build-arm64
```

## Run

```bash
./golan --port 15353 --zone local.lan
```

## Delegate a zone from Dnsmasq

```
server=/local.lan/127.0.0.1#15353
server=/0.168.192.in-addr.arpa/127.0.0.1#15353
```