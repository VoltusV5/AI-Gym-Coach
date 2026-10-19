# Fitness App (Спортик)

## Проверки Go backend

Версия Go указана в `backend/go.mod`, версия golangci-lint — в
`backend/.golangci-version` (2.10.1). Из корня проекта:

```bash
make -C backend lint-install  # установка линтера с проверкой SHA256 релиза
make -C backend fmt           # исправление форматирования
make -C backend check         # gofmt, go vet, golangci-lint и все Go-тесты
make -C backend test-cover build
python3 -m unittest discover -s scripts/tests -v  # сценарии деплоя без сервера
```

CI выполняет те же команды при каждом push и pull request. Отсутствующий линтер
или несовпадение его версии считается ошибкой; автоматической подмены на `go vet` нет.

## Подготовка production-деплоя

[Инструкция CD, настройки сервера и откат](docs/07_DEPLOYMENT.md).
Workflow `.github/workflows/cd.yml` подготовлен, но деплой включается только после
настройки Secrets, окружения `production` и переменной `DEPLOY_ENABLED=true`.
Pull request не запускает деплой. [Отчёт об улучшениях и правках PR #9](REPORT_SPRINT1.md).

## Запуск 

**1. Из корня**

```запуск всех сервисов одним таргетом
make up
```

**2. Запуск по отдельности(перейти в 3-х терминалах в соответствующие папки)**

```для ml, backend, frontend
make sportapp-deploy
make env-up (для postgresql в папке backend)
make env-port-forward (в папке backend, чтобы в pgdmin смотреть бд)
```

**3. Порты**

```
backend - 5050
ml - 8000
frontend - 3000
```

**4. Примечание**

```Docker
У меня запускалось всё только при открытом docker desktop.
Все работает в контейнерах, не падая.
```

## Проблема frontend

```
Сайт в браузере открывается, но там просто белое окно, где ничего нельзя делать. Имитируя frontend запросами на backend через postman у меня с ml все работало правильно во всех смыслах.
```




Поднять весь стек (если ещё не запущен): 
docker compose up -d

Перезапустить без пересборки
docker compose restart

Перезапустить один сервис
docker compose restart frontend (или backend, ml, sportapp-postgres)

Пересобрать и поднять один сервис
docker compose up -d --build frontend

Пересобрать всё сразу
docker compose up -d --build

Остановить и убрать контейнеры
docker compose down
