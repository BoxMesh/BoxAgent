BOXES = {
  "debian" => "bento/debian-13",
  "ubuntu" => "bento/ubuntu-24.04",
  "rocky" => "bento/rockylinux-9",
}

n = Integer(ENV.fetch("N", "1"))
distros = ENV.fetch("DISTROS", "debian,ubuntu,rocky").split(",").map(&:strip)

Vagrant.configure("2") do |config|
  # Vagrant on Windows misdetects the host architecture and then cannot find the amd64 boxes.
  config.vm.box_architecture = "amd64"
  config.vm.synced_folder ".", "/vagrant", disabled: true
  config.vm.boot_timeout = 600

  config.vm.provider "virtualbox" do |vb|
    vb.memory = 512
    vb.cpus = 1
    vb.linked_clone = true
  end

  config.vm.provision "agent-bin", type: "file", run: "never", source: "dist/boxagent", destination: "/tmp/boxagent"
  server = ENV.fetch("BOXSERVER", "192.168.56.1:8081")
  config.vm.provision "agent", type: "shell", run: "never", inline: "install -m 0755 /tmp/boxagent /usr/local/bin/boxagent; pkill -x boxagent; nohup boxagent --server #{server} >>/var/log/boxagent.log 2>&1 </dev/null &"

  distros.product((1..n).to_a).each_with_index do |(distro, i), idx|
    name = "#{distro}-#{i}"
    config.vm.define name do |m|
      m.vm.box = BOXES.fetch(distro)
      m.vm.hostname = name
      m.vm.network "private_network", ip: "192.168.56.#{11 + idx}"
    end
  end
end
