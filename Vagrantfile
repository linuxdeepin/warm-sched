# coding: utf-8
# -*- mode: ruby -*-
# vi: set ft=ruby :

# All Vagrant configuration is done below. The "2" in Vagrant.configure
# configures the configuration version (we support older styles for
# backwards compatibility). Please don't change it unless you know what
# you're doing.
Vagrant.configure("2") do |config|
  config.vm.box = "debian/contrib-stretch64"

  config.vm.synced_folder "./", "/home/vagrant/work", type: "virtualbox"

  config.vm.box_check_update = true

  config.vm.provision "shell", inline: <<-SHELL
    bash -c 'echo "deb http://ftp.cn.debian.org/debian stretch main" > /etc/apt/sources.list'
    apt-get update
    apt-get install -y debhelper golang-go golang-golang-x-sys-dev dpkg-dev linux-headers-amd64
    bash -c 'echo "cd ~/work" >> /home/vagrant/.bashrc'
  SHELL
end
