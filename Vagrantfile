Vagrant.configure("2") do |config|

  config.vm.provider :libvirt do |libvirt|
    libvirt.cpus = 4
    libvirt.memory = 2048
  end

  config.vm.define "master" do |master|
    master.vm.box = "fedora/27-cloud-base"
    master.vm.provision "shell", inline: "hostnamectl set-hostname master"
    master.vm.network "private_network", ip: "192.168.122.10"
  end

  config.vm.define "node1" do |node1|
    node1.vm.box = "fedora/27-cloud-base"
    node1.vm.provision "shell", inline: "hostnamectl set-hostname node1"
  end

  config.vm.define "node2" do |node2|
    node2.vm.box = "fedora/27-cloud-base"
    node2.vm.provision "shell", inline: "hostnamectl set-hostname node2"
  end

end

