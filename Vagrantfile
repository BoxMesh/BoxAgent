# -*- mode: ruby -*-
# N test VMs per distro for BoxAgent. PowerShell:
#   $env:N=3; $env:DISTROS="alpine,debian"; vagrant up
# Keep the same N/DISTROS for later commands (status/ssh/destroy): Vagrant only sees machines defined here.
BOXES = {
  "alpine" => "cloud-image/alpine-3.24", # musl: publish BoxAgent with -r linux-musl-x64
  "debian" => "bento/debian-13",         # glibc: publish with -r linux-x64
}

n = Integer(ENV.fetch("N", "1"))
distros = ENV.fetch("DISTROS", "alpine").split(",").map(&:strip)

Vagrant.configure("2") do |config|
  # Light boxes may lack VirtualBox Guest Additions; deploy with `vagrant upload` instead of /vagrant.
  config.vm.synced_folder ".", "/vagrant", disabled: true

  config.vm.provider "virtualbox" do |vb|
    vb.memory = 512
    vb.cpus = 1
    vb.linked_clone = true # clones share one base disk: faster up, less space
  end

  distros.product((1..n).to_a).each_with_index do |(distro, i), idx|
    name = "#{distro}-#{i}"
    config.vm.define name do |m|
      m.vm.box = BOXES.fetch(distro)
      m.vm.hostname = name
      # VMs reach each other and the host (192.168.56.1) on this network
      m.vm.network "private_network", ip: "192.168.56.#{11 + idx}"
      if distro == "alpine"
        m.ssh.shell = "sh" # Alpine has no bash by default; Vagrant's default is `bash -l`
        m.vm.provision "shell", inline: "apk add --no-cache libgcc libstdc++" # .NET runtime deps
      end
    end
  end
end
