.PHONY: all up backend ml frontend

all: up

up:
	$(MAKE) -C ai sportapp-deploy
	$(MAKE) -C backend env-up
	$(MAKE) -C backend sportapp-deploy
	$(MAKE) -C frontend sportapp-deploy