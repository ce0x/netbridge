# NetBridge Supported Protocols

## Xray Family

### VLESS
```
vless://uuid@server:port?security=tls&sni=example.com&type=ws&path=/path#name
```

### VMess
```
vmess://base64-encoded-json
```

### Trojan
```
trojan://password@server:port?security=tls&sni=example.com&type=ws#name
```

### Shadowsocks
```
ss://base64(method:password)@server:port#name
```

## Native VPN

### WireGuard
```ini
[Interface]
PrivateKey = ...
Address = 10.0.0.2/32

[Peer]
PublicKey = ...
Endpoint = server:51820
AllowedIPs = 0.0.0.0/0
```

### OpenVPN
```
remote server 1194
proto udp
dev tun
```

## Import Sources

- VLESS links
- VMess links
- Trojan links
- Shadowsocks links
- WireGuard config files (.conf)
- OpenVPN config files (.ovpn)
- JSON format
- YAML format
- Subscription URLs
