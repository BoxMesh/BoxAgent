# BoxAgent

Runs on each machine and samples its metrics. For now it prints one JSON line per sample to stdout. Sending samples to BoxServer comes next.

## Layout

```text
cmd/boxagent/        entry point: flags, sampling loop, clean shutdown
internal/metrics/    what gets measured
  snapshot.go        the JSON shape (mirrors BoxMesh/src/types/device.ts)
  collector.go       Collector: machine info, CPU, RAM, disk
  network.go         network rates (Mbps) and network cards
  process.go         busiest processes
deploy/install.sh    installs the agent as a systemd service
Vagrantfile          test VMs: debian-1, ubuntu-1, rocky-1 on 192.168.56.11-13
```

## Commands

| Command | What it does |
|---|---|
| `make test` | Unit tests, plus one test that samples the machine running it |
| `make build` | Static Linux binary in `dist/boxagent` |
| `vagrant up` | Start the 3 VMs (needs `dist/boxagent` built first) |
| `make deploy` | Build, then install or update the agent on every running VM |
| `vagrant ssh debian-1 -c "tail -n1 /var/log/boxagent.log"` | Latest sample from one VM |

## Releases

Every push to `main` runs [release.yml](.github/workflows/release.yml). It runs vet and the tests, then publishes the GitHub release `v0.1.<run number>`. The release holds `boxagent-linux-amd64`, `boxagent-linux-arm64`, `boxagent-windows-amd64.exe` and `SHA256SUMS`.

Flags: `-interval` sets the time between samples (default `10s`). `-top` sets how many processes to report (default `10`).
