# BoxAgent

BoxAgent chạy trên từng máy được quản lý. Cứ vài giây nó đo máy một lần và báo số liệu về **BoxServer** qua gRPC.
BoxMesh, web console, hiển thị những gì BoxServer đã lưu.

```text
BoxAgent (mỗi máy) ──gRPC, chỉ gọi ra ngoài──► BoxServer ◄──HTTPS── BoxMesh (trình duyệt)
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
| Linux, macOS | `/var/lib/boxagent/id` |
| Windows | `%ProgramData%\BoxAgent\id` |

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
| `processes` | repeated | `pid`, `name`, `user`, `status`, `cpu` (%), `memory_mb` |

Định nghĩa nằm ở [proto/device.proto](proto/device.proto). File này phải giống hệt
`BoxServer/src/BoxServer.API/Protos/Device.proto`; sửa thì sửa cả hai, rồi chạy `make proto`.

BoxServer từ chối lệnh gọi có `id` không phải UUID hoặc `host_name` rỗng (`INVALID_ARGUMENT`). Còn lại, nó tạo thiết bị
ở lần gọi đầu và cập nhật ở các lần sau.

### Agent nhận gì

Chưa có gì để agent xử lý. `Ping` trả `true` khi BoxServer đã lưu mẫu, và agent không đọc phản hồi đó. Server hiện chưa
gửi được cho agent lệnh, truy vấn hay cấu hình nào.

BoxMesh đã có màn hình terminal từ xa, duyệt file và web service, nhưng chúng cần kênh này nên chưa chạy được với
BoxServer thật cho tới khi kênh có mặt.

## Cài đặt

Bản build sẵn nằm ở [trang releases](https://github.com/BoxMesh/BoxAgent/releases). Mỗi lần push vào `main` sẽ phát hành
một release `v0.1.<n>` gồm:

| File | Dùng cho |
|---|---|
| `boxagent-linux-amd64` | Linux trên Intel hoặc AMD |
| `boxagent-linux-arm64` | Linux trên ARM (Raspberry Pi, server ARM) |
| `boxagent-windows-amd64.exe` | Windows |
| `SHA256SUMS` | Checksum của ba file trên |

Chưa có bản cho macOS; xem [macOS](#macos).

Trước khi cài, hãy chắc chắn máy tới được BoxServer. Chạy trên máy cần cài:

```bash
nc -zv 192.168.1.10 8081                     # Linux hoặc macOS
Test-NetConnection 192.168.1.10 -Port 8081   # Windows PowerShell
```

### Linux

**Chạy thử bằng tay** (dừng khi đóng shell):

```bash
curl -LO https://github.com/BoxMesh/BoxAgent/releases/latest/download/boxagent-linux-amd64
curl -LO https://github.com/BoxMesh/BoxAgent/releases/latest/download/SHA256SUMS
sha256sum -c SHA256SUMS --ignore-missing      # phải in OK
chmod +x boxagent-linux-amd64
sudo ./boxagent-linux-amd64 --server 192.168.1.10:8081
```

Trên máy ARM, tải `boxagent-linux-arm64` thay cho bản amd64.

**Chạy như service** (tự chạy khi khởi động máy, tự chạy lại khi agent dừng) bằng systemd. Cần quyền root. Đổi
`192.168.1.10:8081` thành địa chỉ gRPC của BoxServer:

```bash
sudo install -m 0755 boxagent-linux-amd64 /usr/local/bin/boxagent

sudo tee /etc/systemd/system/boxagent.service >/dev/null <<'EOF'
[Unit]
Description=BoxAgent
After=network-online.target
Wants=network-online.target

[Service]
ExecStart=/usr/local/bin/boxagent --server 192.168.1.10:8081
Restart=always

[Install]
WantedBy=multi-user.target
EOF

sudo systemctl daemon-reload
sudo systemctl enable --now boxagent
```

| Việc | Lệnh |
|---|---|
| Xem trạng thái | `systemctl status boxagent` |
| Xem log (dòng khởi động và lỗi; mẫu đi thẳng về server) | `journalctl -u boxagent -f` |
| Đổi server | Sửa `--server` trong `/etc/systemd/system/boxagent.service`, rồi `sudo systemctl daemon-reload && sudo systemctl restart boxagent` |
| Nâng cấp | Chép binary mới đè `/usr/local/bin/boxagent`, rồi `sudo systemctl restart boxagent` |
| Gỡ | `sudo systemctl disable --now boxagent && sudo rm /etc/systemd/system/boxagent.service /usr/local/bin/boxagent && sudo systemctl daemon-reload` |

Service chạy bằng root nên đọc được mọi tiến trình. Hệ không có systemd (ví dụ Alpine dùng OpenRC) cần tự viết file
service; trong lúc đó hãy chạy binary bằng tay.

### Windows

Mở PowerShell **bằng quyền Administrator** (agent ghi file ID vào `C:\ProgramData`):

```powershell
mkdir C:\BoxAgent
Invoke-WebRequest https://github.com/BoxMesh/BoxAgent/releases/latest/download/boxagent-windows-amd64.exe -OutFile C:\BoxAgent\boxagent.exe
C:\BoxAgent\boxagent.exe --server 192.168.1.10:8081
```

Chưa có trình cài đặt hay Windows service, nên đóng cửa sổ là agent dừng. Trong lúc chờ, có thể dùng scheduled task để
chạy khi khởi động máy (dự án chưa kiểm chứng cách này):

```powershell
schtasks /Create /TN BoxAgent /SC ONSTART /RU SYSTEM /RL HIGHEST /TR "C:\BoxAgent\boxagent.exe --server 192.168.1.10:8081"
```

Windows Defender Firewall không cần rule inbound, vì agent chỉ kết nối ra ngoài.

### macOS

Chưa phát hành bản cho macOS. Mã nguồn đã xử lý macOS (OS được báo là macOS, ổ đĩa là `/`) nhưng chưa được thử. Hãy tự
build bằng [Go](https://go.dev/dl/) 1.27.3 trở lên:

```bash
git clone https://github.com/BoxMesh/BoxAgent.git && cd BoxAgent
GOOS=darwin GOARCH=arm64 CGO_ENABLED=0 go build -o boxagent ./cmd/boxagent   # Apple silicon; đổi amd64 cho Intel
sudo ./boxagent --server 192.168.1.10:8081
```

Chưa có service launchd.

### Máy ảo để thử

`Vagrantfile` định nghĩa ba VM VirtualBox cho việc phát triển: `debian-1`, `ubuntu-1` và `rocky-1` ở
`192.168.56.11` đến `.13`. Chúng tới BoxServer đang chạy trên máy host ở `192.168.56.1:8081`.

```bash
make build                       # binary Linux tĩnh ở dist/boxagent
vagrant up                       # bật các VM
make deploy                      # chạy hoặc cập nhật agent trên mọi VM đang chạy
vagrant ssh debian-1 -c "tail -n1 /var/log/boxagent.log"
```

Đặt `BOXSERVER=host:port` để trỏ sang server khác, `DISTROS=debian,ubuntu` để chọn bản phân phối, và `N=2` để mỗi loại
có nhiều hơn một VM.

## Build từ mã nguồn

Cần Go 1.27.3 trở lên.

| Lệnh | Tác dụng |
|---|---|
| `make test` | Chạy unit test, kèm một test lấy mẫu chính máy đang chạy |
| `make build` | Binary Linux amd64 tĩnh ở `dist/boxagent`, phiên bản `0.1.<số commit>` |
| `make proto` | Sinh lại `internal/pb` sau khi sửa proto (cần `protoc`, `protoc-gen-go`, `protoc-gen-go-grpc`) |
| `make deploy` | Build rồi chạy hoặc cập nhật agent trên mọi VM thử đang chạy |

Nền tảng khác: `GOOS=windows GOARCH=amd64 go build -ldflags "-s -w -X main.version=0.1.99" -o boxagent.exe ./cmd/boxagent`.
Thiếu cờ `-X main.version=…` thì agent báo phiên bản `0.0.0`.

