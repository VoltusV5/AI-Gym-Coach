# 06. Тесты — Postman

## Импорт и запуск

1. В Postman нажмите **Import** и загрузите [коллекцию](postman/Sportik.postman_collection.json) и [окружение](postman/Local.postman_environment.json).
2. Выберите окружение «Спортик · Локальная тестовая БД». Запустите backend с отдельной тестовой БД; заполните `base_url` и `admin_token` без `Bearer`, установите `test_environment=true`.
3. Укажите `max_stage` от 0 до 15 — последний проверяемый этап (безопасный начальный default4). Go-этапы реализованы до 15; для этапа 5 нужен доступный ML v2 либо явно настроенный тестовый ответ. Для этапов 07+ заполните `verification_document_ref` ссылкой на тестовый документ тренера.
4. Нажмите **Run collection**: порядок сверху вниз, **Iterations = 1**, **Keep variable values** включён. Результаты — в **Tests**. После исправлений повторите запуск.

Коллекция сама создаёт аккаунты и основные данные, сохраняет токены и ID. Данные остаются в тестовой БД. Для сценариев оплаты нужен `PAYMENTS_MODE=stub`; реальных списаний нет. Четыре интеграционных шаблона папки 16 выключены по умолчанию и не считаются успешными при пропуске.

## Воспроизводимые проверки репозитория

```sh
GOTOOLCHAIN=go1.25.5 make -C backend check
npm --prefix frontend ci
npm --prefix frontend run check
PLAYWRIGHT_CHANNEL=chrome npm --prefix frontend run test:e2e
```

Последняя команда использует установленный Chrome. Для bundled Chromium сначала выполните `npx playwright install chromium` из frontend и не задавайте PLAYWRIGHT_CHANNEL. Браузерные тесты поднимают локальный production preview на 127.0.0.1:5179. Обычный `test:e2e` проверяет управляемые ответы/ошибки API; это отдельно от реальной API-интеграции.

Полный Go-сценарий требует новую пустую БД с префиксом `sportik_sprint_test_` на отдельном PostgreSQL127.0.0.1:55443, роль sportik_test. Используйте отдельный тестовый кластер; не берите рабочую `.env`, не очищайте существующую БД. Тест отказывается удалять существующую sportapp schema.

```sh
SPORTIK_SPRINT_TEST_DB=sportik_sprint_test_unique_run GOTOOLCHAIN=go1.25.5 go -C backend test ./internal/features -run TestSprintIntegration -count=1
```

`SPORTIK_BROWSER_TESTS=1` дополнительно запускает desktop/mobile Playwright с реальным Go API, синтетическим JWT и одноразовыми фикстурами. `SPORTIK_POSTMAN_TESTS=1` исполняет коллекцию через Newman на том же изолированном API. Для второго режима установите отдельные инструменты в игнорируемый каталог:

```sh
npm install --prefix .validation/postman-tools --no-save --ignore-scripts chai@4 ajv@6 newman@6
```

Флаги можно совместить. Каждый повтор использует новую БД; тест сам не удаляет данные. Временный файл с синтетическими тестовыми токенами создаётся с правами 0600 и удаляется после прогона. Newman-отчёт не сохраняет токены/пароли и находится в `.validation/postman-real-report.json`.

Пересборка документации и коллекции:

```sh
python3 docs/api/build_openapi.py
python3 docs/postman/build_collection.py
```

`docs/postman/verify_collection.cjs` проверяет JavaScript, схемы и чувствительность утверждений на синтетических ответах. Это не замена настоящему Newman-прогону. [Зафиксированные результаты и ограничения](IMPLEMENTATION_REPORT.md).

## Ограничения Postman

Раньше бесплатный Postman предоставлял **25 запусков коллекции в месяц**. С марта 2026 на новых тарифах, включая **Free**, запуски безлимитные; старые тарифы могут сохранять лимит. [Актуальные условия](https://learning.postman.com/v11/docs/billing/about-plans#unlimited-collection-runs).

Если ваш аккаунт ограничен 25 запусками, можно перейти на тариф с безлимитным Runner либо отправлять запросы по одному через **Send** и смотреть **Test Results**.
