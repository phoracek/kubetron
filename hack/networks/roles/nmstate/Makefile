default:


DESTDIR ?=

ifeq ($(USER),root)
ANSIBLE_PARENTPATH ?= /usr/share/
else
ANSIBLE_PARENTPATH ?= $(HOME)/.
endif

ANSIBLE_BASE = $(DESTDIR)$(ANSIBLE_PARENTPATH)ansible
ANSIBLE_MODULE = $(ANSIBLE_BASE)/plugins/modules
ANSIBLE_MODULE_UTILS = $(ANSIBLE_BASE)/plugins/module_utils
install:
	install -d "$(ANSIBLE_MODULE)"
	install -d "$(ANSIBLE_MODULE_UTILS)"
	install library/*.py  "$(ANSIBLE_MODULE)"
	install module_utils/*py  "$(ANSIBLE_MODULE_UTILS)"
