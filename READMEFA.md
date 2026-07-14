# NetBridge — راهنمای فارسی

ابزار یکپارچه دسترسی و اتصال شبکه برای سرورهای لینوکس.

---

## فهرست مطالب

1. [معرفی پروژه](#معرفی-پروژه)
2. [پیش‌نیازها و ابزارهای مورد نیاز](#پیش‌نیازها-و-ابزارهای-مورد-نیاز)
3. [نحوه نصب و اجرا](#نحوه-نصب-و-اجرا)
4. [نحوه ساخت (Build) از منبع](#نحوه-ساخت-build-از-منبع)
5. [دستورات کامل](#دستورات-کامل)
6. [پروتکل‌های پشتیبانی شده](#پروتکل‌های-پشتیبانی-شده)
7. [حالت‌های اتصال](#حالت‌های-اتصال)
8. [مدیریت DNS](#مدیریت-dns)
9. [مسیریابی هوشمند](#مسیریابی-هوشمند)
10. [سیستم بهداشت و Failover](#سیستم-بهداشت-و-failover)
11. [Benchmark و امتیازدهی](#benchmark-و-امتیازدهی)
12. [حالت TUI (رابط متنی تعاملی)](#حالت-tui-رابط-متنی-تعاملی)
13. [API خروجی JSON](#api-خروجی-json)
14. [مدیریت سرویس systemd](#مدیریت-سرویس-systemd)
15. [امنیت](#امنیت)
16. [ساختار پروژه](#ساختار-پروژه)
17. [توسعه افزونه (Plugin)](#توسعه-افزونه-plugin)
18. [ایده‌های به‌روزرسانی بعدی](#ایده‌های-به‌روزرسانی-بعدی)
19. [عیب‌یابی](#عیب‌یابی)
20. [لینک‌های مفید](#لینک‌های-مفید)

---

## معرفی پروژه

NetBridge یک ابزار خط فرمان (CLI) و رابط متنی تعاملی (TUI) است که مدیریت اتصالات شبکه را روی سرورهای لینوکس ساده می‌کند. این ابزار از پروتکل‌های مختلف VPN و پروکسی پشتیبانی می‌کند و قابلیت‌های مسیریابی هوشمند، پایش سلامت اتصال، و امتیازدهی خودکار را ارائه می‌دهد.

### ویژگی‌های کلیدی

- پشتیبانی از چندین پروتکل: VLESS, VMess, Trojan, Shadowsocks, WireGuard, OpenVPN
- سه حالت رابط: خط فرمان (CLI)، رابط متنی تعاملی (TUI)، خروجی JSON
- موتور مسیریابی هوشمند
- پایش سلامت اتصال و Failover خودکار
- Benchmark و امتیازدهی به پروفایل‌ها
- مدیریت DNS
- SDK افزونه برای توسعه‌پذیری

---

## پیش‌نیازها و ابزارهای مورد نیاز

### ابزارهای ضروری

| ابزار | نسخه حداقل | توضیح |
|-------|-----------|-------|
| **Go** | 1.22+ | زبان برنامه‌نویسی اصلی پروژه — برای کامپایل و اجرا الزامی است |
| **Git** | هر نسخه | برای کلون کردن مخزن و مدیریت نسخه |
| **Make** | هر نسخه | اجرای دستورات ساخت و نصب |
| **systemd** | — | مدیریت سرویس (اختیاری برای نصب به صورت سرویس) |

### ابزارهای پشتیبان (Backend)

این ابزارها توسط NetBridge مدیریت می‌شوند و برای اتصال واقعی به سرور نیاز هستند:

| ابزار | کاربرد | محل دریافت |
|-------|--------|-----------|
| **Xray Core** | اجرای پروتکل‌های VLESS, VMess, Trojan, Shadowsocks | [github.com/XTLS/Xray-core](https://github.com/XTLS/Xray-core) |
| **sing-box** | جایگزین Xray برای پروتکل‌های مشابه | [github.com/SagerNet/sing-box](https://github.com/SagerNet/sing-box) |
| **WireGuard** | اجرای پروتکل WireGuard | بسته `wireguard-tools` در اکثر توزیع‌ها |
| **OpenVPN** | اجرای پروتکل OpenVPN | بسته `openvpn` در اکثر توزیع‌ها |

### ابزارهای توسعه (اختیاری)

| ابزار | کاربرد |
|-------|--------|
| **golangci-lint** | بررسی کیفیت کد با `make lint` |
| **GoReleaser** | ساخت نسخه‌های رسمی و انتشار |
| **Docker** | تست در محیط‌های ایزوله |

### نصب Go در لینوکس

```bash
# دانلود و نصب Go 1.22+
wget https://go.dev/dl/go1.22.4.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.22.4.linux-amd64.tar.gz

# اضافه کردن به PATH
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# بررسی نصب
go version
```

### نصب Git

```bash
# Debian/Ubuntu
sudo apt update && sudo apt install git -y

# CentOS/RHEL
sudo yum install git -y

# Arch Linux
sudo pacman -S git
```

---

## نحوه نصب و اجرا

### نصب سریع (یک خطی)

```bash
curl -fsSL https://get.netbridge.dev | bash
```

### نصب دستی

```bash
# کلون کردن مخزن
git clone https://github.com/netbridge/netbridge
cd netbridge

# ساخت و نصب
make build
sudo make install
```

### اجرای سریع

```bash
# وارد کردن یک پروفایل
netbridge import "vless://uuid@server:443?security=tls#my-server"

# نمایش لیست پروفایل‌ها
netbridge list

# اتصال
netbridge connect my-server

# بررسی وضعیت
netbridge status
```

---

## نحوه ساخت (Build) از منبع

### دستورات Make

```bash
# ساخت باینری
make build

# اجرای تست‌ها
make test

# بررسی کیفیت کد
make lint

# قالب‌بندی کد
make vet

# نصب باینری در /usr/local/bin
sudo make install

# پاکسازی فایل‌های ساخت
make clean

# اجرای تمام مراحل (fmt + vet + test + build)
make all

# اجرای مستقیم بدون ساخت باینری
make dev
```

### ساخت دستی با Go

```bash
# ساخت با Go
CGO_ENABLED=0 go build -ldflags "-s -w -X main.version=dev" -o build/netbridge ./cmd/

# اجرا
./build/netbridge --help
```

### ساخت نسخه با GoReleaser

```bash
# نصب GoReleaser
go install github.com/goreleaser/goreleaser@latest

# ساخت نسخه snapshot
goreleaser release --snapshot --clean
```

---

## دستورات کامل

### مدیریت پروفایل

| دستور | توضیح |
|-------|-------|
| `netbridge import <url\|file>` | وارد کردن پروفایل از URL یا فایل |
| `netbridge export <name>` | خروجی گرفتن از پروفایل |
| `netbridge delete <name>` | حذف پروفایل |
| `netbridge rename <old> <new>` | تغییر نام پروفایل |
| `netbridge clone <name> <new>` | کپی پروفایل |
| `netbridge list` | نمایش لیست تمام پروفایل‌ها |
| `netbridge use <name>` | تنظیم پروفایل فعال |
| `netbridge show <name>` | نمایش جزئیات پروفایل |

### مدیریت اتصال

| دستور | توضیح |
|-------|-------|
| `netbridge connect [profile]` | اتصال با پروفایل فعال یا مشخص |
| `netbridge disconnect` | قطع اتصال فعلی |
| `netbridge restart` | راه‌اندازی مجدد اتصال |
| `netbridge reload` | بازخوانی تنظیمات بدون قطع اتصال |
| `netbridge status` | نمایش وضعیت اتصال فعلی |

### حالت‌های اتصال

| دستور | توضیح |
|-------|-------|
| `netbridge connect --mode socks` | پروکسی محلی SOCKS5 در `127.0.0.1:10808` |
| `netbridge connect --mode http` | پروکسی محلی HTTP در `127.0.0.1:8080` |
| `netbridge connect --mode tun` | رابط مجازی TUN |
| `netbridge run <command>` | اجرای دستور از طریق پروفایل فعال |

### تست و سلامت

| دستور | توضیح |
|-------|-------|
| `netbridge test [profile]` | تست اتصال پروفایل |
| `netbridge health [profile]` | بررسی سلامت پروفایل |
| `netbridge benchmark [--all]` | Benchmark و امتیازدهی به پروفایل‌ها |

### مسیریابی هوشمند

| دستور | توضیح |
|-------|-------|
| `netbridge route add <domain> <profile>` | اضافه کردن قانون مسیریابی |
| `netbridge route remove <domain>` | حذف قانون مسیریابی |
| `netbridge route list` | نمایش لیست قوانین مسیریابی |
| `netbridge route clear` | پاک کردن تمام قوانین |

### مدیریت DNS

| دستور | توضیح |
|-------|-------|
| `netbridge dns list` | نمایش پیش‌فرض‌های DNS |
| `netbridge dns use <preset\|ip>` | تنظیم رفع‌کننده DNS فعال |
| `netbridge dns benchmark` | Benchmark رفع‌کننده‌های DNS |
| `netbridge dns show` | نمایش تنظیمات DNS فعلی |
| `netbridge dns reset` | بازگردانی DNS به حالت پیش‌فرض سیستم |

### ادغام با Shell

| دستور | توضیح |
|-------|-------|
| `netbridge env` | چاپ متغیرهای محیطی پروکسی |
| `netbridge unset` | چاپ دستور حذف متغیرهای پروکسی |

### مدیریت سرویس

| دستور | توضیح |
|-------|-------|
| `netbridge service install` | نصب واحد systemd |
| `netbridge service start` | شروع سرویس |
| `netbridge service stop` | توقف سرویس |
| `netbridge service restart` | راه‌اندازی مجدد سرویس |
| `netbridge service status` | نمایش وضعیت سرویس |
| `netbridge service uninstall` | حذف واحد systemd |

### پلاگین‌ها

| دستور | توضیح |
|-------|-------|
| `netbridge plugin load <path>` | بارگذاری افزونه |
| `netbridge plugin list` | نمایش لیست افزونه‌ها |
| `netbridge plugin unload <name>` | تخلیه افزونه |

### پرچم‌های سراسری

| پرچم | توضیح |
|------|-------|
| `--json` | خروجی به فرمت JSON |
| `-q, --quiet` | سرکوب خروجی غیرضروری |
| `-v, --verbose` | فعال‌سازی خروجی مفصل |

---

## پروتکل‌های پشتیبانی شده

### خانواده Xray

| پروتکل | فرمت لینک |
|--------|-----------|
| **VLESS** | `vless://uuid@server:port?security=tls&sni=example.com&type=ws&path=/path#name` |
| **VMess** | `vmess://base64-encoded-json` |
| **Trojan** | `trojan://password@server:port?security=tls&sni=example.com&type=ws#name` |
| **Shadowsocks** | `ss://base64(method:password)@server:port#name` |

### VPN بومی

| پروتکل | فرمت فایل |
|--------|-----------|
| **WireGuard** | فایل `.conf` با ساختار `[Interface]` و `[Peer]` |
| **OpenVPN** | فایل `.ovpn` با تنظیمات `remote`, `proto`, `dev` |

### منابع وارد کردن

- لینک‌های VLESS, VMess, Trojan, Shadowsocks
- فایل‌های پیکربندی WireGuard (.conf)
- فایل‌های پیکربندی OpenVPN (.ovpn)
- فرمت JSON
- فرمت YAML
- آدرس URL اشتراک (Subscription)

---

## حالت‌های اتصال

### SOCKS5

```bash
netbridge connect --mode socks
# آدرس محلی: 127.0.0.1:10808
```

### HTTP Proxy

```bash
netbridge connect --mode http
# آدرس محلی: 127.0.0.1:8080
```

### TUN

```bash
netbridge connect --mode tun
# ایجاد رابط مجازی TUN
```

### استفاده متغیرهای محیطی

```bash
# فعال‌سازی پروکسی در shell فعلی
eval $(netbridge env)

# غیرفعال‌سازی پروکسی
eval $(netbridge unset)
```

---

## مدیریت DNS

### پیش‌فرض‌های DNS موجود

```bash
netbridge dns list
```

### تنظیم DNS

```bash
# استفاده از یک پیش‌فرض
netbridge dns use cloudflare

# استفاده از آدرس IP
netbridge dns use 8.8.8.8
```

### Benchmark DNS

```bash
netbridge dns benchmark
# مقایسه سرعت و کیفیت رفع‌کننده‌های مختلف DNS
```

---

## مسیریابی هوشمند

مسیریابی هوشمند امکان هدایت ترافیک دامنه‌های خاص از طریق پروفایل‌های مختلف را فراهم می‌کند.

```bash
# اضافه کردن قانون
netbridge route add example.com my-server

# نمایش قوانین
netbridge route list

# حذف قانون
netbridge route remove example.com

# پاک کردن تمام قوانین
netbridge route clear
```

---

## سیستم بهداشت و Failover

### بررسی سلامت

```bash
netbridge health my-server
# نمایش: دسترسی، تاخیر، از دست دادن بسته‌ها
```

### Failover خودکار

زنجیره Failover امکان جابجایی خودکار به پروفایل بعدی در صورت قطع اتصال را فراهم می‌کند:

```bash
netbridge failover add chain1 profile1,profile2,profile3
netbridge failover start
```

---

## Benchmark و امتیازدهی

```bash
# Benchmark یک پروفایل
netbridge benchmark my-server

# Benchmark تمام پروفایل‌ها
netbridge benchmark --all

# معیارهای امتیازدهی:
# - تاخیر (Latency)
# - نوسان (Jitter)
# - پهنای باند (Throughput)
# - از دست دادن بسته‌ها (Packet Loss)
```

---

## حالت TUI (رابط متنی تعاملی)

```bash
netbridge tui
```

### بخش‌های TUI

- **منوی اصلی**: دسترسی سریع به تمام امکانات
- **اتصال**: مدیریت اتصال فعلی
- **پروفایل‌ها**: مدیریت لیست پروفایل‌ها
- **مسیریابی**: مدیریت قوانین مسیریابی
- **DNS**: مدیریت تنظیمات DNS
- **آمار**: نمایش آمار ترافیک
- **Benchmark**: اجرای Benchmark
- **لاگ‌ها**: مشاهده لاگ‌ها
- **تنظیمات**: تغییر تنظیمات

---

## API خروجی JSON

تمام دستورات از پرچم `--json` پشتیبانی می‌کنند:

```bash
netbridge status --json
netbridge list --json
netbridge benchmark --all --json
netbridge dns list --json
```

### نمونه خروجی JSON

```json
{
  "status": "connected",
  "profile": "my-server",
  "mode": "socks",
  "local_addr": "127.0.0.1:10808",
  "bytes_up": 1024000,
  "bytes_down": 5120000,
  "uptime": "2h30m"
}
```

---

## مدیریت سرویس systemd

### نصب سرویس

```bash
# نصب خودکار با ساخت فایل systemd
sudo netbridge service install

# یا نصب دستی
sudo cp systemd/netbridge.service /etc/systemd/system/
sudo systemctl daemon-reload
sudo systemctl enable netbridge
```

### مدیریت سرویس

```bash
sudo netbridge service start     # شروع
sudo netbridge service stop      # توقف
sudo netbridge service restart   # راه‌اندازی مجدد
sudo netbridge service status    # بررسی وضعیت
sudo netbridge service uninstall # حذف
```

### استفاده مستقیم از systemd

```bash
sudo systemctl start netbridge
sudo systemctl status netbridge
sudo journalctl -u netbridge -f   # مشاهده لاگ‌ها به صورت زنده
```

---

## امنیت

### رمزنگاری

- تمام فایل‌های پروفایل با **AES-256-GCM** رمزنگاری می‌شوند
- کلید اصلی از شناسه ماشین مشتق می‌شود
- کلید فقط در حافظه نگهداری می‌شود و هرگز روی دیسک نوشته نمی‌شود

### مجوزهای فایل

| مسیر | مجوز |
|------|------|
| دایرکتوری تنظیمات | `700` |
| فایل‌های پروفایل | `600` |
| فایل‌های لاگ | `640` |

### پوشاندن اطلاعات حساس

- رمزها در لاگ‌ها و خروجی پوشانده می‌شوند
- کلیدها در خروجی `netbridge show` پوشانده می‌شوند
- استفاده از پرچم `--reveal` برای نمایش مقادیر حساس

### امنیت شبکه

- تأیید TLS به صورت پیش‌فرض فعال است
- پرچم `--allow-insecure` فقط برای تست استفاده شود
- پشتیبانی از پروتکل Reality برای مبهم‌سازی پیشرفته

### توصیه‌های امنیتی

1. با حداقل دسترسی‌ها اجرا کنید
2. از عبارت عبور قوی برای رمزنگاری استفاده کنید
3. به صورت دوره‌ای اعتبارنامه‌ها را تغییر دهید
4. لاگ‌ها را برای دسترسی‌های غیرمجاز پایش کنید
5. NetBridge را به‌روز نگه دارید

---

## ساختار پروژه

```
NetBridge/
├── cmd/                    # دستورات CLI (cobra)
│   ├── root.go             # دستور اصلی
│   ├── main.go             # نقطه ورود
│   ├── connect.go          # اتصال
│   ├── disconnect.go       # قطع اتصال
│   ├── status.go           # وضعیت
│   ├── profile.go          # مدیریت پروفایل
│   ├── benchmark.go        # Benchmark
│   ├── dns.go              # مدیریت DNS
│   ├── route.go            # مسیریابی
│   ├── tui.go              # رابط متنی تعاملی
│   ├── service.go          # مدیریت سرویس
│   ├── plugin.go           # مدیریت افزونه
│   └── ...
├── internal/               # لایه تجاری
│   ├── core/               # موتور اصلی
│   ├── config/             # پیکربندی و بارگذاری
│   ├── profile/            # مدیریت پروفایل
│   ├── session/            # مدیریت نشست
│   ├── routing/            # موتور مسیریابی
│   ├── health/             # پایش سلامت و Failover
│   ├── benchmark/          # موتور Benchmark
│   ├── dns/                # موتور DNS
│   ├── stats/              # جمع‌آوری آمار
│   └── security/           # رمزنگاری و امنیت
├── adapters/               # آداپتورهای بک‌اند
│   ├── xray/               # Xray Core
│   ├── singbox/            # sing-box
│   ├── wireguard/          # WireGuard
│   └── openvpn/            # OpenVPN
├── plugins/                # سیستم افزونه
│   ├── sdk.go              # SDK افزونه
│   ├── registry.go         # رجیستری افزونه
│   └── examples/           # نمونه افزونه‌ها
├── tui/                    # رابط متنی تعاملی
│   ├── app.go              # برنامه اصلی TUI
│   ├── views/              # نماها
│   │   ├── connect.go      # نمای اتصال
│   │   ├── profiles.go     # نمای پروفایل‌ها
│   │   ├── routing.go      # نمای مسیریابی
│   │   ├── dns.go          # نمای DNS
│   │   ├── stats.go        # نمای آمار
│   │   ├── benchmark.go    # نمای Benchmark
│   │   ├── logs.go         # نمای لاگ‌ها
│   │   └── settings.go     # نمای تنظیمات
│   └── styles.go           # استایل‌ها
├── pkg/                    # بسته‌های عمومی
│   ├── uri/                # پارسر URI
│   ├── netutil/            # ابزارهای شبکه
│   ├── humanize/           # قالب‌بندی انسانی
│   ├── jsonout/            # خروجی JSON
│   └── sysinfo/            # اطلاعات سیستم
├── scripts/                # اسکریپت‌ها
│   ├── build.sh            # اسکریپت ساخت
│   ├── install.sh          # اسکریپت نصب
│   ├── dev-setup.sh        # راه‌اندازی محیط توسعه
│   └── release.sh          # اسکریپت انتشار
├── docs/                   # مستندات
├── tests/                  # تست‌ها
├── systemd/                # فایل‌های systemd
├── go.mod                  # ماژول Go
├── Makefile                # دستورات ساخت
└── interfaces.go           # قراردادهای اصلی
```

---

## توسعه افزونه (Plugin)

### ایجاد افزونه

1. یک دایرکتوری جدید در `plugins/` ایجاد کنید
2. رابط `Plugin` را پیاده‌سازی کنید
3. افزونه را در رجیستری ثبت کنید

### نمونه افزونه

```go
package myplugin

import netbridge "github.com/netbridge/netbridge"

type MyPlugin struct{}

func (p *MyPlugin) Name() string { return "myplugin" }
func (p *MyPlugin) Version() string { return "1.0.0" }
func (p *MyPlugin) Protocols() []netbridge.Protocol {
    return []netbridge.Protocol{"myproto"}
}
func (p *MyPlugin) NewBackend(profile netbridge.Profile) (netbridge.Backend, error) {
    return &MyBackend{}, nil
}
```

### بارگذاری افزونه

```bash
netbridge plugin load ./plugins/myplugin
netbridge plugin list
netbridge plugin unload myplugin
```

---

## ایده‌های به‌روزرسانی بعدی

### ویژگی‌های پیشنهادی

| اولویت | ویژگی | توضیح |
|--------|-------|-------|
| بالا | **پشتیبانی از TUIC** | اضافه کردن پروتکل TUIC به عنوان بک‌اند جدید |
| بالا | **پشتیبانی از Hysteria** | اضافه کردن پروتکل Hysteria 2 |
| بالا | **پشتیبانی از SSH Tunnel** | اتصال از طریق SSH به عنوان تونل |
| متوسط | **پشتیبانی از Cloudflare Tunnel** | اتصال از طریق Cloudflare Zero Trust |
| متوسط | **رابط گرافیکی (GUI)** | ایجاد رابط گرافیکی با استفاده از Fyne یا Wails |
| متوسط | **داشبورد وب** | رابط وب برای مدیریت از مرورگر |
| متوسط | **اعلان‌های دسکتاپ** | اطلاع‌رسانی وضعیت اتصال |
| متوسط | **پشتیبانی از macOS و Windows** | گسترش پشتیبانی به سیستم‌عامل‌های دیگر |
| پایین | **آمار و نمودار** | نمودارهای گرافیکی مصرف ترافیک |
| پایین | **سیستم اعلان قطعی** | اطلاع‌رسانی خودکار در صورت قطع اتصال |
| پایین | **پشتیبانی از Docker** | اجرای NetBridge در کانتینر Docker |
| پایین | **CLI تعاملی هوشمند** | تکمیل خودکار دستورات |

### بهبودهای فنی

| اولویت | بهبود | توضیح |
|--------|-------|-------|
| بالا | **بهبود تست‌ها** | افزایش پوشش تست به بالای 80% |
| بالا | **مستندسازی کامل** | مستندسازی تمام API‌های داخلی |
| متوسط | **بهینه‌سازی عملکرد** | بهبود سرعت Benchmark و مسیریابی |
| متوسط | **مدیریت خطا بهتر** | پیام‌های خطای واضح‌تر و راهنمایی بیشتر |
| متوسط | **پشتیبانی از فایل پیکربندی TOML** | اضافه کردن فرمت TOML در کنار YAML و JSON |
| پایین | **کش هوشمند** | کش کردن نتایج Benchmark و DNS |
| پایین | **سیستم لاگ بهتر** | لاگ‌های ساختاریافته با فرمت JSON |

### ایده‌های جامعه

- پشتیبانی از پروتکل‌های جدید با توجه به نیازهای جامعه
- ترجمه مستندات به زبان‌های مختلف
- ایجاد پکیج‌های APT/YUM/DNF برای نصب آسان
- پشتیبانی از ARM برای اجرا روی Raspberry Pi

---

## عیب‌یابی

### خطاهای رایج

| خطا | راه حل |
|-----|--------|
| `profile not found` | پروفایل وجود ندارد. با `netbridge list` بررسی کنید |
| `no active session` | اتصالی فعال نیست. ابتدا `netbridge connect` را اجرا کنید |
| `already connected` | قبلاً متصل هستید. ابتدا `netbridge disconnect` اجرا کنید |
| `permission denied` | نیاز به دسترسی root. با `sudo` اجرا کنید |
| `backend not found` | بک‌اند پروتکل پیدا نشد. Xray یا sing-box را نصب کنید |

### بررسی لاگ‌ها

```bash
# لاگ‌های سرویس
sudo journalctl -u netbridge -f

# لاگ‌های NetBridge
ls /etc/netbridge/logs/
cat /etc/netbridge/logs/netbridge.log
```

### تست اتصال

```bash
# تست یک پروفایل خاص
netbridge test my-server

# بررسی سلامت
netbridge health my-server

# نمایش وضعیت فعلی
netbridge status
```

---

## لینک‌های مفید

- **مخزن اصلی**: https://github.com/netbridge/netbridge
- ** مستندات**: [docs/](docs/)
- **Issues**: https://github.com/netbridge/netbridge/issues
- **Go**: https://golang.org/dl/
- **Xray Core**: https://github.com/XTLS/Xray-core
- **sing-box**: https://github.com/SagerNet/sing-box
- **WireGuard**: https://www.wireguard.com/

---

## مجوز

MIT License
