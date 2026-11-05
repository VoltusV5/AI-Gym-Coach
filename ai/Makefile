-include .env
export

sportapp-deploy:
	@docker compose up -d --build ml

.PHONY: test
test:
	@PYTHONDONTWRITEBYTECODE=1 python -m unittest discover -s tests -v
