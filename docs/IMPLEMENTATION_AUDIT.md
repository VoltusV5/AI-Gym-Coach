# Повторная проверка Jira: фронтенд и Go backend

Дата: 14.11.2026. Источник — экспорт 82 карточек доски [KAN / Спортик-v2 разработка](https://voltusv-dev.atlassian.net/jira/software/projects/KAN/boards/2), полученный в этой сессии. Отобраны 29 BE, 22 FE и 2 QA; всего **53 задачи**. Карточки с уже существовавшей реализацией проверялись по коду и критериям, а не по статусу backlog. Jira не изменялась.

Для каждой строки сверены «Что сделать» и критерии исходной карточки. «Реализовано» ниже означает наличие поведения в согласованных границах и указанную проверку; не означает 100% branch coverage или отдельный браузерный тест каждого элемента UI. Требование Jira о новых комментариях в коде отменено прямым указанием владельца. Платёжный провайдер/выплаты — согласованная заглушка; ML ranking/explain — Go-интеграция и fallback, без реализации ML.

## Матрица требований

| Jira | Задача | Состояние и свидетельство |
|---|---|---|
| [KAN-21](https://voltusv-dev.atlassian.net/browse/KAN-21) | BE-01 | Реализовано. Слои HTTP/service/repository, ошибки, конфигурация, v1; backend check и legacy-тесты. |
| [KAN-22](https://voltusv-dev.atlassian.net/browse/KAN-22) | FE-01 | Реализовано. Общий layout/errors/client; исправлены также видимые ошибки старых onboarding шагов и повторная отправка; lint/build/mock браузер. |
| [KAN-23](https://voltusv-dev.atlassian.net/browse/KAN-23) | BE-02 | Реализовано. Unit, синтетические fixtures, TestSprintIntegration и Postman; без внешнего AI и рабочей БД. |
| [KAN-24](https://voltusv-dev.atlassian.net/browse/KAN-24) | FE-02 | Реализовано. Vitest/component/Playwright; session isolation/editor/draft и новые generation/review E2E. |
| [KAN-25](https://voltusv-dev.atlassian.net/browse/KAN-25) | BE-03 | Реализовано. JWT, AccountStatus, роли/ownership, валидация, limit/cursor; отрицательные API-сценарии. |
| [KAN-26](https://voltusv-dev.atlassian.net/browse/KAN-26) | FE-03 | Реализовано. api session revision, role guards, busy/error/409; исправлена пагинация с сохранением применённых фильтров; unit. |
| [KAN-32](https://voltusv-dev.atlassian.net/browse/KAN-32) | BE-04 | Реализовано. catalog/reference seed; добавлен sessions_per_week 1–7 через HTTP/service/SQL/курсор; synthetic SQL-регрессии. |
| [KAN-33](https://voltusv-dev.atlassian.net/browse/KAN-33) | BE-05 | Реализовано. catalog/programs, карточка, entitlement; скрытие draft и неизменность опубликованной версии проверены API. |
| [KAN-34](https://voltusv-dev.atlassian.net/browse/KAN-34) | FE-04 | Реализовано. Catalog/Program, filters/reset/cursor/access; добавлена частота и полный equipment selector; mock + real браузер. |
| [KAN-35](https://voltusv-dev.atlassian.net/browse/KAN-35) | BE-06 | Реализовано. favorites: повтор, удаление, чужой пользователь и скрытая программа; Go API + браузер. |
| [KAN-36](https://voltusv-dev.atlassian.net/browse/KAN-36) | FE-05 | Реализовано. Favorites/Program: подтверждённые сервером save/unsave, обновление при входе; real браузер. |
| [KAN-37](https://voltusv-dev.atlassian.net/browse/KAN-37) | BE-07 | Реализовано. profile: optimistic version, nullable поля, equipment/days; generation constraints + Postman. |
| [KAN-38](https://voltusv-dev.atlassian.net/browse/KAN-38) | FE-06 | Реализовано. Profile/onboarding: nullable/version/conflict, все страницы equipment; mock conflict и реальный save→generation. |
| [KAN-39](https://voltusv-dev.atlassian.net/browse/KAN-39) | BE-08 | Реализовано. access: self-admin запреты, роли, blocked и audit before/after; конкурентные/отрицательные Go API проверки. |
| [KAN-40](https://voltusv-dev.atlassian.net/browse/KAN-40) | FE-07 | Реализовано. Users: protected admin route, search/cursor, reason/confirmation/reload; API негативные сценарии, route guard в браузере. |
| [KAN-41](https://voltusv-dev.atlassian.net/browse/KAN-41) | BE-09 | Реализовано. plans: CRUD/revisions/schedule/activate/archive/idempotency; Go API и редактор в браузере. |
| [KAN-42](https://voltusv-dev.atlassian.net/browse/KAN-42) | FE-08 | Реализовано. TrainingDaysEditor/ScheduleEditor/PlanEditor: copy/reorder/load/save/activate/archive; mock failed-save и real manual/AI редактор. |
| [KAN-43](https://voltusv-dev.atlassian.net/browse/KAN-43) | BE-10 | Реализовано. generation: version/idempotency/deadline/ошибки ML, atomic save; добавлена Go-проверка safe_for_injuries и отсутствие записи при отказе. |
| [KAN-44](https://voltusv-dev.atlassian.net/browse/KAN-44) | FE-09 | Реализовано. Profile/PlanGenerating: stable idempotency key, полнота профиля, abort/re-entry; реальный questionnaire→generation→editor→save→reload E2E. |
| [KAN-45](https://voltusv-dev.atlassian.net/browse/KAN-45) | BE-11 | Реализовано. workouts: фактические подходы/revision/provenance/idempotency; конкурентные повторы, ownership и браузер. |
| [KAN-46](https://voltusv-dev.atlassian.net/browse/KAN-46) | BE-12 | Реализовано. workouts/domain: агрегаты/рекорды/adherence/streak/вес/периоды; unit с известными результатами и SQL-интеграции. |
| [KAN-47](https://voltusv-dev.atlassian.net/browse/KAN-47) | FE-10 | Реализовано. Progress/Workout/ExerciseProgress: history, aggregates/chart/table/records/weight/units; real training→progress браузер. |
| [KAN-48](https://voltusv-dev.atlassian.net/browse/KAN-48) | BE-13 | Реализовано. trainers: public/own DTO, приватные материалы заявки; закрытые материалы не попадают в публичный API. |
| [KAN-49](https://voltusv-dev.atlassian.net/browse/KAN-49) | BE-14 | Реализовано. trainers: admin decision/reason/audit, конкурентное решение и неверные роли; synthetic API. |
| [KAN-50](https://voltusv-dev.atlassian.net/browse/KAN-50) | FE-11 | Реализовано. Trainer/Profile/Verification: public/private state, materials, admin decisions; код UI сверён с API, серверный путь покрыт интеграцией. |
| [KAN-51](https://voltusv-dev.atlassian.net/browse/KAN-51) | BE-15 | Реализовано. publications: draft/version, новые IDs и независимость опубликованной версии; synthetic API + Postman. |
| [KAN-52](https://voltusv-dev.atlassian.net/browse/KAN-52) | BE-16 | Реализовано. publications: submit/moderate/published catalog, причины/конфликты; synthetic API + Postman. |
| [KAN-53](https://voltusv-dev.atlassian.net/browse/KAN-53) | FE-12 | Реализовано. AuthorEditor/Programs/Moderation: draft/version/submit, категории без усечения; UI сверён с контрактом, полный серверный цикл в Go/Postman. |
| [KAN-54](https://voltusv-dev.atlassian.net/browse/KAN-54) | BE-17 | Реализовано. commerce: согласованный sandbox, HMAC/replay, price/commission snapshot, ledger/entitlement; API + реальная браузерная покупка. |
| [KAN-55](https://voltusv-dev.atlassian.net/browse/KAN-55) | FE-13 | Реализовано. Checkout/Orders: server state, stub label, failed/pay/fork/refund; реальные desktop/mobile E2E. |
| [KAN-56](https://voltusv-dev.atlassian.net/browse/KAN-56) | BE-18 | Реализовано. commerce: trainer balance/payout, commission/effective_at, refunds/paid_at; synthetic API и Postman. Перевод денег — заглушка. |
| [KAN-57](https://voltusv-dev.atlassian.net/browse/KAN-57) | FE-14 | Реализовано. Finance: trainer/admin payout/history/commission и реальный server reload; UI контракт + financial API, mock period regression. |
| [KAN-58](https://voltusv-dev.atlassian.net/browse/KAN-58) | BE-19 | Реализовано. sharing: независимый Fork, entitlement и source version, новые IDs; SQL/API и реальный браузер. |
| [KAN-59](https://voltusv-dev.atlassian.net/browse/KAN-59) | BE-20 | Реализовано. sharing: immutable snapshot/token hash/expiry/revoke/license/library; SQL/API и реальный браузер. |
| [KAN-60](https://voltusv-dev.atlassian.net/browse/KAN-60) | FE-15 | Реализовано. ShareControls/PublicShare/Library: source/copy, preview/save/fork/revoke; реальные desktop/mobile E2E. |
| [KAN-61](https://voltusv-dev.atlassian.net/browse/KAN-61) | BE-21 | Реализовано. social: порог 3 занятия, доступ/автор/повторы/рейтинги; исправлена версия на фактически использованную, включая восстановление удалённого отзыва. |
| [KAN-62](https://voltusv-dev.atlassian.net/browse/KAN-62) | BE-22 | Реализовано. social: reports/moderation/reason, hidden/deleted исключаются из рейтинга; target_review только admin; API + mock браузер. |
| [KAN-63](https://voltusv-dev.atlassian.net/browse/KAN-63) | FE-16 | Реализовано. Reviews/Reports: eligibility/count/errors/own review и admin target; новый real 0→3 workouts→create→reload→delete E2E. |
| [KAN-64](https://voltusv-dev.atlassian.net/browse/KAN-64) | BE-23 | Реализовано. social/outbox: follow/unfollow, уведомления о публикации, owner/read и дедупликация; synthetic API. |
| [KAN-65](https://voltusv-dev.atlassian.net/browse/KAN-65) | FE-17 | Реализовано. Trainers/Notifications: server follow/read/target; UI контракт + Go follow/publication/outbox интеграция. |
| [KAN-66](https://voltusv-dev.atlassian.net/browse/KAN-66) | BE-24 | Реализовано. social/events: whitelist, JWT user, hash/UUID dedup, запрещённые поля; synthetic API + Postman. |
| [KAN-67](https://voltusv-dev.atlassian.net/browse/KAN-67) | FE-18 | Реализовано. ProgramCard/events: visible impressions, view/save/recommendation_click, real IDs; API whitelist/dedup проверены Go/Postman. |
| [KAN-68](https://voltusv-dev.atlassian.net/browse/KAN-68) | BE-25 | Реализовано. insights/mlclient: rank/explain validation/deadline/recheck, deterministic fallback; unit/API. Реализация ML не изменялась. |
| [KAN-69](https://voltusv-dev.atlassian.net/browse/KAN-69) | FE-19 | Реализовано. Recommendations: server explanations/fallback notice, score без обещания результата; real браузер + rank unit/API. |
| [KAN-70](https://voltusv-dev.atlassian.net/browse/KAN-70) | BE-26 | Реализовано. insights/dashboard: приоритеты/даты/состояния; отдельные SQL SAVEPOINT для блоков, реальные ошибки чтения не отключают главную. |
| [KAN-71](https://voltusv-dev.atlassian.net/browse/KAN-71) | FE-20 | Реализовано. Dashboard: server block priorities/empty/unavailable/actions; real браузер, отказы блоков дополнительно проверены Go SQL. |
| [KAN-72](https://voltusv-dev.atlassian.net/browse/KAN-72) | BE-27 | Реализовано. insights/analytics: trainer/admin роли/периоды/финансы/paid_at; unit и synthetic API. |
| [KAN-73](https://voltusv-dev.atlassian.net/browse/KAN-73) | FE-21 | Реализовано. Admin navigation: users/trainers/programs/references/reviews/reports/orders/finance/analytics; роли/аналитика API + браузерные guard/moderation. |
| [KAN-74](https://voltusv-dev.atlassian.net/browse/KAN-74) | BE-28 | Реализовано. references: nullable CRUD без hard delete, причины/audit, inactive и seed override; SQL-регрессии повторного seed/конкурентной очистки. |
| [KAN-75](https://voltusv-dev.atlassian.net/browse/KAN-75) | FE-22 | Реализовано. References: paged lists/form/reason/nullable media/inactive; полный equipment selector; admin media regression в браузере и seed SQL. |
| [KAN-76](https://voltusv-dev.atlassian.net/browse/KAN-76) | BE-29 | Реализовано. insights snapshot: repeatable read профиля/истории/поведения/кандидатов; только schema-поля собственных оценок, без текста отзывов. |
| [KAN-77](https://voltusv-dev.atlassian.net/browse/KAN-77) | QA-01 | Проверено в доступном окружении. Изолированный full Go/SQL/Newman; up/down/up миграций, отрицательные сценарии, OpenAPI/Postman/schema и legacy preservation. |
| [KAN-78](https://voltusv-dev.atlassian.net/browse/KAN-78) | QA-02 | Проверено в доступном окружении. Web production build, desktop Chrome и iPhone viewport Chrome. Native APK/IPA и Docker image build не подтверждены: нет native SDK/projects и доступного Docker daemon; ограничение допустимо критерием QA-02 и перечислено явно. |

## Исправления после повторного аудита

1. В каталоге отсутствовала фильтрация по частоте: добавлен `sessions_per_week` (1–7), включая UI, SQL, область курсора, OpenAPI и Postman.
2. Отзыв ошибочно получал текущую опубликованную версию: теперь выбирается последняя версия из фактически завершённых занятий; при редактировании привязка сохраняется, при восстановлении удалённого отзыва определяется заново.
3. Ошибка запроса истории/подписок отключала всю главную: блоки читаются в одном repeatable-read снимке с отдельными SAVEPOINT, ошибочный блок получает `unavailable`. Потеря соединения/общий deadline остаются ошибкой запроса.
4. Go-генератор дополнительно проверяет безопасность упражнения при `injuries_notes=true` непосредственно перед atomic save, даже если ML вернул неподходящее упражнение.
5. Ошибки сохранения старой анкеты теперь видны в форме; введённые поля сохраняются для повтора.
6. Следующая страница использует снимок успешно применённых фильтров, а не редактируемые поля. Справочные селекторы проходят все страницы, включая категории после сотой.
7. Генератор создаёт новый AbortController для попытки и не применяет запоздалый ответ после ухода/повторного входа.
8. Добавлены браузерные критические цепочки анкета→генерация→редактор→сохранение и запрет отзыва→3 настоящих занятия→отзыв→удаление.

## Проверки и границы

Результаты окончательного прогона фиксируются в [IMPLEMENTATION_REPORT.md](IMPLEMENTATION_REPORT.md); команды — в [06_TESTS.md](06_TESTS.md). Независимые read-only ревью исходных критериев и итогового diff выполнены отдельно для Go и фронтенда. Замечания исправлены.

Git scope сверяется с базой `decf5a0`: изменения только в `frontend/`, `backend/`, вспомогательной документации/контрактах/тестах `docs/` и `.gitignore`. `ai/`, ML/Python runtime, CI и compose не изменены. Python-файлы в `docs/` — генераторы документации и тестовой коллекции; это не сервисный backend.

Рабочая ветка `codex/sportik-v2-frontend-go-backlog`; push не выполнялся. Подробные решения для команды — в игнорируемом `TEAM_IMPLEMENTATION_NOTES.md`.
