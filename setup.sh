#!/bin/bash

set -e

# install terraform
# https://learn.hashicorp.com/tutorials/terraform/install-cli
sudo yum install -y yum-utils
sudo yum-config-manager --add-repo https://rpm.releases.hashicorp.com/AmazonLinux/hashicorp.repo
sudo yum -y install terraform

# install golang
amazon-linux-extras install golang1.11

# get required module
go get github.com/hashicorp/terraform-exec/tfexec
go get github.com/hashicorp/terraform-exec/tfinstall
