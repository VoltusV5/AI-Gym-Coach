# 02. API — Swagger

Swagger — интерактивная документация API. Полный контракт Go и ML: [openapi.yml](api/openapi.yml).

## Как использовать

- **Онлайн:** откройте [Swagger Editor](https://editor.swagger.io/), выберите **File → Import file** и загрузите `docs/api/openapi.yml`.
- **Локально:** из корня проекта `AI-Gym-Coach` выполните `python docs/api/serve.py` и откройте [Swagger проекта](http://localhost:8088/api/). В «Новый семестр» доступны API для **Backend / Frontend** и **ML**; эти методы пока не реализованы.

Дополнительно:

- **Запросы:** запустите `docker compose up -d` в `AI-Gym-Coach`; для Go получите `token` через `auth/guest` или `auth/login` и вставьте в **Authorize** без `Bearer`.
- **Обновление:** измените `docs/api/openapi.json`, затем выполните `python docs/api/build_openapi.py` — обновятся локальные разделы и единый YAML.

- **Тесты в Postman:** импортируйте [коллекцию](postman/Sportik.postman_collection.json) и [окружение](postman/Local.postman_environment.json); заполните `base_url`, `admin_token`, выберите `max_stage` и запустите Runner. [Инструкция](06_TESTS.md).

<!-- Сохранены якоря ссылок из остальных документов. -->
<a id="f01"></a><a id="f02"></a><a id="f03"></a><a id="f04"></a><a id="f05"></a><a id="f06"></a><a id="f07"></a><a id="f08"></a><a id="f09"></a><a id="f10"></a><a id="f11"></a><a id="f12"></a>
