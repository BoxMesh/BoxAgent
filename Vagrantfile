BOXES = {
  "alpine" => "cloud-image/alpine-3.24",
  "debian" => "bento/debian-13",
}

n = Integer(ENV.fetch("N", "1"))
distros = ENV.fetch("DISTROS", "alpine").split(",").map(&:strip)

Vagrant.configure("2") do |config|
  config.vm.synced_folder ".", "/vagrant", disabled: true

  config.vm.provider "virtualbox" do |vb|
    vb.memory = 512
    vb.cpus = 1
    vb.linked_clone = true
  end

  distros.product((1..n).to_a).each_with_index do |(distro, i), idx|
    name = "#{distro}-#{i}"
    config.vm.define name do |m|
      m.vm.box = BOXES.fetch(distro)
      m.vm.hostname = name
      m.vm.network "private_network", ip: "192.168.56.#{11 + idx}"
      if distro == "alpine"
        m.ssh.shell = "sh"
        m.vm.provision "shell", inline: "apk add --no-cache libgcc libstdc++"
      end
    end
  end
end
