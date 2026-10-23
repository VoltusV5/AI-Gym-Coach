---
name: create-endpoint
description: Создание или изменение HTTP endpoint в Go backend Спортика с существующими слоями, JWT, валидацией и тестами контракта.
---

# Endpoint в Go backend

1. Прочитать корневой и `backend/AGENTS.md`, контракт задачи и соседний endpoint. Для v1 свериться с `docs/api/current.json`; не переносить планируемые функции v2 в v1 автоматически.
2. Определить метод/путь, JSON, ошибки, гостевой доступ, владельца объекта и пагинацию. Если меняется контракт, согласовать его до реализации. Не расширять разрешённую пользователем область правок.
3. В `internal/features/<feature>/transport/http/` декодировать запрос через `DecodeAndValidateRequest`, получить пользователя из контекста, вызвать сервис и ответить через `NewHTTPResponce`. Зарегистрировать защищённый маршрут с `Protect(jwt)`; текущий префикс `/api/v1` добавляет общий router.
4. Оставить бизнес-правила в service, SQL в repository. Добавить метод в нужный интерфейс и существующие моки. SQL ограничить владельцем, учитывать deleted_at и version; использовать таймаут пула, закрытие rows и проверку `rows.Err()`.
5. Добавить GoDoc и целевые тесты: успех, неверный ввод, отсутствие/невалидный JWT через маршрут с middleware, чужой объект, применимые 404/409 и ошибка зависимости. Использовать моки/`httptest`, без рабочей БД и внешнего ML.
6. При изменении API обновить OpenAPI/Postman, если это входит в область задачи; иначе описать необходимое согласование. Запустить целевой `go test`, `make -C backend verify`, затем установленный `make -C backend lint`. Передать diff отдельному субагенту на ревью и обработать замечания по общему процессу.

## Пример из проекта

`backend/internal/features/users/transport/http/transport.go` регистрирует удаление заметки:

```go
core_http_server.NewRoute(
    http.MethodDelete, "/notes/{id}", h.DeleteNote, protect,
)
```

Цепочка: `transport/http/notes.go: DeleteNote` → `service/notes.go: DeleteNotesUser` → `repository/postgres/notes.go: DeleteNotesUser`. Транспорт берёт `userID` из JWT-контекста; SQL использует `WHERE id = $1 AND user_id = $2 AND deleted_at IS NULL`, а ноль затронутых строк означает `ErrNotFound`. Успех — 204. Проверка контракта: `backend/internal/features/users/transport/http/transport_test.go`, тест `TestUsersHTTP_Notes_PaginationAndIDOR`.

Из корня:

```sh
(cd backend && go test ./internal/features/users/transport/http -run TestUsersHTTP_Notes_PaginationAndIDOR -count=1)
make -C backend verify
```

Готовность: маршрут зарегистрирован, права и ошибки проверены, контракт сохранён либо согласован, проверки и независимое ревью описаны в результате.
