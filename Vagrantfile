BOXES = {
  "debian" => "bento/debian-13",
  "ubuntu" => "bento/ubuntu-24.04",
  "rocky" => "bento/rockylinux-9",
}

n = Integer(ENV.fetch("N", "1"))
distros = ENV.fetch("DISTROS", "debian,ubuntu,rocky").split(",").map(&:strip)

Vagrant.configure("2") do |config|
  config.vm.synced_folder ".", "/vagrant", disabled: true
  # With Hyper-V on, VirtualBox runs through it (NEM) and boots slowly; the 300s default is too tight.
  config.vm.boot_timeout = 600

  config.vm.provider "virtualbox" do |vb|
    vb.memory = 512
    vb.cpus = 1
    vb.linked_clone = true
  end

  # Install or update the agent with `make deploy`; `vagrant up` alone never runs these.
  config.vm.provision "agent-bin", type: "file", run: "never", source: "dist/boxagent", destination: "/tmp/boxagent"
  config.vm.provision "agent-install", type: "file", run: "never", source: "deploy/install.sh", destination: "/tmp/install.sh"
  # The VMs reach the Windows host, where BoxServer runs, at its host-only address.
  server = ENV.fetch("BOXSERVER", "192.168.56.1:8081")
  config.vm.provision "agent", type: "shell", run: "never", inline: "sh /tmp/install.sh #{server}"

  distros.product((1..n).to_a).each_with_index do |(distro, i), idx|
    name = "#{distro}-#{i}"
    config.vm.define name do |m|
      m.vm.box = BOXES.fetch(distro)
      m.vm.hostname = name
      m.vm.network "private_network", ip: "192.168.56.#{11 + idx}"
    end
  end
end
