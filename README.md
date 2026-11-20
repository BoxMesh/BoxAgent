# ZExporter

ZExporter chạy trên từng máy được quản lý. Cứ vài giây nó đo máy một lần và báo số liệu về **BoxServer** qua gRPC.
Zin, web console, hiển thị những gì BoxServer đã lưu.

```text
ZExporter (mỗi máy) ──gRPC, chỉ gọi ra ngoài──► BoxServer ◄──HTTPS── Zin (trình duyệt)
```

Agent chỉ gọi ra (dial-out). Máy không cần public IP, không cần mở inbound port, không cần VPN; chỉ cần đường tới cổng
gRPC của BoxServer (mặc định 8081).

## Hiện tại agent làm được gì

| Đã làm | Chưa làm |
|---|---|
| Đo CPU, RAM, ổ đĩa, mạng, uptime và toàn bộ tiến trình | Nhận bất kỳ lệnh hay yêu cầu nào từ BoxServer |
| Báo từng mẫu về BoxServer (RPC `Ping`) | Terminal, desktop, truyền file, service từ xa |
| Giữ định danh thiết bị cố định qua các lần khởi động lại | Tự xác thực (enrollment token) |
| | Mã hoá kết nối (TLS) |

Xem thêm [Lệnh và yêu cầu](#lệnh-và-yêu-cầu).

## Agent thu thập gì

Một **mẫu** được lấy mỗi `--interval` (mặc định 10 giây). Các tốc độ được tính trên khoảng thời gian từ mẫu trước.

### Mức sử dụng

| Chỉ số | Đơn vị | Cách đo |
|---|---|---|
| CPU | phần trăm, 0–100 | Thời gian bận của cả máy kể từ mẫu trước |
| RAM | phần trăm | Bộ nhớ đã dùng so với tổng |
| Disk | phần trăm | Dung lượng đã dùng của **một** volume: `/` trên Linux và macOS, `C:\` trên Windows |
| Mạng vào / ra | Mbps | Số byte nhận và gửi của tất cả card mạng cộng lại, kể từ mẫu trước. Bộ đếm bị lùi (reset hoặc tràn) tính là 0 |

### Thông tin máy

| Trường | Ví dụ | Ghi chú |
|---|---|---|
| Hostname | `debian-1` | Lấy từ hệ điều hành. Có thể trùng giữa các máy nên không dùng làm định danh |
| Tên và phiên bản OS | `debian`, `13` | Theo platform và platform version của gopsutil |
| Kiến trúc | `x86_64` | Kiến trúc kernel |
| Tên CPU | `AMD EPYC 7313P 16-Core Processor` | Đọc một lần rồi giữ lại |
| Tổng RAM | byte | |
| Tổng ổ đĩa | byte | Cùng một volume với phần trăm Disk |
| Uptime | giây | |
| IP và MAC | `192.168.56.12`, `08:00:27:…` | Của card mạng mà hệ điều hành sẽ dùng để tới BoxServer. VM có hai card sẽ báo card mà server tới được. Để trống nếu không xác định được đường đi |
| Phiên bản agent | `0.1.14` | Gắn lúc build; `go build` thường cho `0.0.0` |

### Tiến trình

Agent luôn gửi **toàn bộ** tiến trình đang chạy, không giới hạn số lượng, xếp theo CPU rồi tới bộ nhớ (bận nhất trước).

| Trường | Ghi chú |
|---|---|
| PID, tên, người dùng | |
| PPID | PID của tiến trình cha, `0` khi không có hoặc không đọc được. Console dùng nó để gom các tiến trình cùng tên thành một dòng ứng dụng |
| Trạng thái | `running`, `suspended` hoặc `not-responding` (zombie) |
| CPU | Phần trăm của **cả máy** (tiến trình dùng hết một nhân trên máy 4 nhân thì ra 25) |
| Bộ nhớ | Bộ nhớ resident (RSS) tính bằng MB |

Tiến trình mà agent không đủ quyền đọc sẽ bị bỏ qua. Chạy agent bằng root hoặc Administrator để thấy đủ.

### Mẫu đầu tiên

Tốc độ mạng và CPU theo tiến trình cần hai lần đọc, nên ở mẫu đầu tiên sau khi agent khởi động chúng bằng 0.

## Định danh thiết bị

Ở lần chạy đầu, agent sinh một UUID ngẫu nhiên và lưu vào file. BoxServer dùng nó để nhận ra máy, nên máy giữ nguyên
lịch sử và alias sau khi khởi động lại hay đổi hostname.

| Hệ điều hành | File |
|---|---|
| Linux, macOS | `/var/lib/zexporter/id` |
| Windows | `%ProgramData%\ZExporter\id` |

- File được tạo với quyền 0600 và chỉ chứa UUID.
- File chứa nội dung khác thì agent dừng và báo lỗi, không tự ghi đè, vì ID mới sẽ hiện thành một thiết bị mới.
- Muốn một máy hiện như thiết bị mới thì xoá file này. Khi nhân bản VM, hãy xoá file ở bản sao, nếu không hai máy sẽ báo
  cáo như một.

## Lệnh và yêu cầu

### Dòng lệnh

Không có lệnh con (dùng [cobra](https://github.com/spf13/cobra)); agent nhận các cờ, viết với hai gạch (`--server`):

| Cờ | Mặc định | Ý nghĩa |
|---|---|---|
| `-s`, `--server host:port` | bắt buộc | Địa chỉ **gRPC** của BoxServer, ví dụ `192.168.1.10:8081`. Thiếu cờ này agent báo lỗi và thoát |
| `-i`, `--interval` | `10s` | Thời gian giữa hai mẫu, kiểu duration của Go: `5s`, `30s`, `1m` |
| `-v`, `--version` | | In phiên bản agent rồi thoát |
| `-h`, `--help` | | In hướng dẫn dùng |

`SIGINT` hoặc `SIGTERM` (Ctrl+C, `kill`) dừng agent một cách gọn gàng.

### Agent gửi gì

Một lệnh gọi gRPC, `device.DeviceGRPC/Ping`, mỗi mẫu một lần. Lần gọi lỗi hoặc chậm (giới hạn 5 giây) chỉ được ghi log
rồi bỏ; mẫu sau thử lại, và gRPC tự nối lại. Việc lấy mẫu không bao giờ bị kẹt vì server đang tắt.

| Trường | Kiểu | Ý nghĩa |
|---|---|---|
| `id` | string | UUID của thiết bị |
| `host_name` | string | Hostname |
| `os_name`, `os_version`, `os_architecture` | string | Hệ điều hành |
| `ip_address`, `mac_address` | string | Card mạng tới được server |
| `cpu` | string | Tên CPU |
| `ram`, `disk` | int64 | Tổng, tính bằng byte |
| `uptime` | int64 | Giây |
| `agent_version` | string | |
| `cpu_usage`, `ram_usage`, `disk_usage` | double | Phần trăm |
| `net_in`, `net_out` | double | Mbps |
| `processes` | repeated | `pid`, `name`, `user`, `status`, `cpu` (%), `memory_mb`, `ppid` |

Định nghĩa nằm ở [proto/device.proto](proto/device.proto). File này phải giống hệt
`OctopusServer/src/Zin.API/Protos/Device.proto`; sửa thì sửa cả hai, rồi chạy `make proto`.

BoxServer từ chối lệnh gọi có `id` không phải UUID hoặc `host_name` rỗng (`INVALID_ARGUMENT`). Còn lại, nó tạo thiết bị
ở lần gọi đầu và cập nhật ở các lần sau.

### Agent nhận gì

Chưa có gì để agent xử lý. `Ping` trả `true` khi BoxServer đã lưu mẫu, và agent không đọc phản hồi đó. Server hiện chưa
gửi được cho agent lệnh, truy vấn hay cấu hình nào.

Zin đã có màn hình terminal từ xa, duyệt file và web service, nhưng chúng cần kênh này nên chưa chạy được với
BoxServer thật cho tới khi kênh có mặt.

## Cài đặt

Bản build sẵn nằm ở [trang releases](https://github.com/ZinOctopus/exporter/releases). Mỗi lần push vào `main` sẽ phát hành
một release `v0.1.<n>` gồm:

| File | Dùng cho |
|---|---|
| `zexporter-linux-amd64` | Linux trên Intel hoặc AMD |
| `zexporter-linux-arm64` | Linux trên ARM (Raspberry Pi, server ARM) |
| `zexporter-windows.exe` | Windows (trình cài đặt, xem mục Windows) |
| `SHA256SUMS` | Checksum của hai file Linux |

Chưa có bản cho macOS; xem [macOS](#macos).

Trước khi cài, hãy chắc chắn máy tới được BoxServer. Chạy trên máy cần cài:

```bash
nc -zv 192.168.1.10 8081                     # Linux hoặc macOS
Test-NetConnection 192.168.1.10 -Port 8081   # Windows PowerShell
```

### Linux

**Cài bằng script (Ubuntu, Debian và các bản có systemd).** Script tự chọn bản amd64 hoặc arm64, kiểm tra checksum, cài
binary vào `/usr/local/bin/zexporter`, tạo service `zexporter` (tự chạy khi khởi động máy, tự chạy lại khi lỗi) và khởi động nó:

```bash
curl -fsSL https://github.com/ZinOctopus/exporter/releases/latest/download/install.sh | sudo sh -s -- --server 192.168.1.10:8081
```

Chạy lại đúng lệnh đó để nâng cấp hoặc đổi server (không truyền `--server` thì dùng lại địa chỉ đã lưu). Gỡ:
`curl -fsSL https://github.com/ZinOctopus/exporter/releases/latest/download/install.sh | sudo sh -s -- --uninstall`.
Mã ID thiết bị ở `/var/lib/zexporter` được giữ lại. Phần dưới là cách làm tay, dùng khi máy không có systemd hoặc không
muốn chạy script.

**Chạy thử bằng tay** (dừng khi đóng shell):

```bash
curl -LO https://github.com/ZinOctopus/exporter/releases/latest/download/zexporter-linux-amd64
curl -LO https://github.com/ZinOctopus/exporter/releases/latest/download/SHA256SUMS
sha256sum -c SHA256SUMS --ignore-missing      # phải in OK
chmod +x zexporter-linux-amd64
sudo ./zexporter-linux-amd64 --server 192.168.1.10:8081
```

Trên máy ARM, tải `zexporter-linux-arm64` thay cho bản amd64.

**Chạy như service** (tự chạy khi khởi động máy, tự chạy lại khi agent dừng) bằng systemd. Cần quyền root. Đổi
`192.168.1.10:8081` thành địa chỉ gRPC của BoxServer:

```bash
sudo install -m 0755 zexporter-linux-amd64 /usr/local/bin/zexporter

sudo tee /etc/systemd/system/zexporter.service >/dev/null <<'EOF'
[Unit]
Description=ZExporter
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/zexporter --server 192.168.1.10:8081
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now zexporter
```

| Việc | Lệnh |
|---|---|
| Xem trạng thái | `systemctl status zexporter` |
| Xem log (dòng khởi động và lỗi; mẫu đi thẳng về server) | `journalctl -u zexporter -f` |
| Đổi server | Sửa `--server` trong `/etc/systemd/system/zexporter.service`, rồi `sudo systemctl daemon-reload && sudo systemctl restart zexporter` |
| Nâng cấp | Chép binary mới đè `/usr/local/bin/zexporter`, rồi `sudo systemctl restart zexporter` |
| Gỡ | `sudo systemctl disable --now zexporter && sudo rm /etc/systemd/system/zexporter.service /usr/local/bin/zexporter && sudo systemctl daemon-reload` |

Service chạy bằng root nên đọc được mọi tiến trình. Hệ không có systemd (ví dụ Alpine dùng OpenRC) cần tự viết file
service; trong lúc đó hãy chạy binary bằng tay.

### Windows

Tải `zexporter-windows.exe` ở trang Releases và chạy (cần quyền Administrator). Đây là trình cài đặt: nhập địa chỉ
`host:port` của server, nó chép agent vào `C:\Program Files\ZExporter`, tạo service `ZExporter` (tự chạy khi khởi động máy, tự
chạy lại khi lỗi) và khởi động nó. Service chạy bằng `LocalSystem`, nên đọc được mọi tiến trình.

- **Đổi địa chỉ hoặc sửa lỗi cài đặt:** chạy lại chính file đó, ô địa chỉ được điền sẵn giá trị cũ.
- **Cài im lặng:** `zexporter-windows.exe /VERYSILENT /SERVER=host:port`.
- **Gỡ:** Settings → Apps. Mã ID thiết bị ở `C:\ProgramData\ZExporter` được giữ lại, nên cài lại vẫn là thiết bị cũ.
- Log của service không được lưu ở đâu cả; khi cần gỡ lỗi, build và chạy bằng tay (xem mục build).

Windows Defender Firewall không cần rule inbound, vì agent chỉ kết nối ra ngoài.

### macOS

Chưa phát hành bản cho macOS. Mã nguồn đã xử lý macOS (OS được báo là macOS, ổ đĩa là `/`) nhưng chưa được thử. Hãy tự
build bằng [Go](https://go.dev/dl/) 1.27.3 trở lên:

```bash
git clone https://github.com/ZinOctopus/exporter.git && cd ZExporter
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o zexporter ./cmd/zexporter   # Apple silicon; đổi amd64 cho Intel
sudo ./zexporter --server 192.168.1.10:8081
```

Chưa có service launchd.

### Máy ảo để thử

`Vagrantfile` định nghĩa ba VM VirtualBox cho việc phát triển: `debian-1`, `ubuntu-1` và `rocky-1` ở
`192.168.56.11` đến `.13`. Chúng tới BoxServer đang chạy trên máy host ở `192.168.56.1:8081`.

```bash
make build                       # binary Linux tĩnh ở dist/zexporter
vagrant up                       # bật các VM
make deploy                      # chạy hoặc cập nhật agent trên mọi VM đang chạy
vagrant ssh debian-1 -c "tail -n1 /var/log/zexporter.log"
```

Đặt `BOXSERVER=host:port` để trỏ sang server khác, `DISTROS=debian,ubuntu` để chọn bản phân phối, và `N=2` để mỗi loại
có nhiều hơn một VM.

## Build từ mã nguồn

Cần Go 1.27.3 trở lên.

| Lệnh | Tác dụng |
|---|---|
| `make test` | Chạy unit test, kèm một test lấy mẫu chính máy đang chạy |
| `make build` | Binary Linux amd64 tĩnh ở `dist/zexporter`, phiên bản `0.1.<số commit>` |
| `make proto` | Sinh lại `internal/pb` sau khi sửa proto (cần `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`) |
| `make deploy` | Build rồi chạy hoặc cập nhật agent trên mọi VM thử đang chạy |

Nền tảng khác: `GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=0.1.99" -o zexporter.exe ./cmd/zexporter`.
Thiếu cờ `-X main.version=…` thì agent báo phiên bản `0.0.0`.

