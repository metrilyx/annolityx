
SHELL = /bin/bash

UNAME = $(shell uname)

## Release file to determin distro and os
UBUNTU_REL_F := /etc/lsb-release
DEBIAN_REL_F := /etc/debian_version
ORACLE_REL_F := /etc/oracle-release
REDHAT_REL_F := /etc/redhat-release
## Determine OS and Distribution
ifeq ($(UNAME), Darwin)
	DISTRO := osx
## Check oracle first as it also has the redhat-release file
else ifneq ("$(wildcard $(ORACLE_REL_F))", "")
	DISTRO := oracle
else ifneq ("$(wildcard $(REDHAT_REL_F))", "")
	DISTRO := redhat
## Check ubuntu first as it also has the debian_version file
else ifneq ("$(wildcard $(UBUNTU_REL_F))","")
	DISTRO := ubuntu
	CODENAME := `grep 'DISTRIB_CODENAME=' $(UBUNTU_REL_F) | cut -d '=' -f 2 | tr '[:upper:]' '[:lower:]'`
else ifneq ("$(wildcard $(DEBIAN_REL_F))", "")
	DISTRO := debian
	CODENAME := `cat $(DEBIAN_REL_F)`
endif

ifeq ("$(DISTRO)", "")
	echo "Could not determine distro!"
	exit 1
endif

START_DATETIME := `date`

NAME := annolityx

EPEL_REPO_RPM := epel-release-6-8.noarch.rpm
EPEL_REPO_URL := http://mirrors.nl.eu.kernel.org/fedora-epel/6/x86_64/$(EPEL_REPO_RPM)
RPM_DEPS := zeromq zeromq-devel
DEB_DEPS := libzmq3-dev

GO_PKG_PATH := github.com/metrilyx/$(NAME)

DATA_DIR := /usr/local/share/$(NAME)

BUILDDIR := $(shell pwd)/build
BUILDROOT := $(BUILDDIR)/$(NAME)

.deps:
	if [[ ( "$(DISTRO)" == "ubuntu" ) || ( "$(DISTRO)" == "debian" ) ]]; then \
		apt-get update -qq; \
		for pkg in $(DEB_DEPS); do \
			apt-get install -q -y $$pkg; \
		done; \
	else \
		yum -y install $(EPEL_REPO_URL); \
		for pkg in $(RPM_DEPS); do \
			yum -y install $$pkg; \
		done; \
	fi;

.build:
	go get -d -v ./...

.test:
	go test -v ./...

.clean:
	rm -rf ./build

install:
	go install -v ./...

	if [ -e "$(BUILDROOT)" ]; then rm -rf "$(BUILDROOT)"; fi;
	
	mkdir -p $(BUILDROOT)/usr/local/bin/;
	cp ../../../../bin/annolityx $(BUILDROOT)/usr/local/bin/

	cp -a etc/annolityx $(BUILDROOT)/etc/annolityx

	mkdir -p $(BUILDROOT)/$(DATA_DIR)	
	cp -a webroot $(BUILDROOT)/$(DATA_DIR)/

	mkdir -p $(BUILDROOT)/$(DATA_DIR)/docs
	cp README.md $(BUILDROOT)/$(DATA_DIR)/docs/

	cd $(BUILDDIR) && tar -czf $(NAME)-$(DISTRO).tgz $(NAME) && cd -
