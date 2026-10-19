#!/bin/sh
# Выполняется на сервере. Секреты читает только Compose из shared/.env.production.
set -eu
action=${1:-}
release=${2:-}
archive=${3:-}
root=${DEPLOY_ROOT:-/srv/sportik}
case "$action" in deploy|rollback) ;; *) echo "Использование: deploy.sh deploy|rollback SHA [archive.tar.gz]" >&2; exit 1 ;; esac
case "$release" in ''|*[!a-f0-9]*) echo "Нужен полный SHA коммита" >&2; exit 1 ;; esac
test "${#release}" -eq 40 || { echo "Нужен SHA длиной 40 символов" >&2; exit 1; }
test -f "$root/shared/.env.production" || { echo "Не настроено окружение сервера" >&2; exit 1; }
mkdir -p "$root/releases"
exec 9>"$root/deploy.lock"
flock -n 9 || { echo "Другой деплой уже выполняется" >&2; exit 1; }

compose() {
  RELEASE_ID="$release" GO_VERSION="$go_version" docker compose \
    --project-name sportik --env-file "$root/shared/.env.production" \
    -f "$root/releases/$release/docker-compose.production.yml" "$@"
}
select_release() {
  test -f "$root/releases/$release/docker-compose.production.yml"
  go_version=$(awk '$1 == "go" {print $2; exit}' "$root/releases/$release/backend/go.mod")
  test -n "$go_version"
}
health() {
  compose exec -T frontend wget -q -O /dev/null http://127.0.0.1:3000/healthz
  compose exec -T frontend wget -q -O /dev/null http://127.0.0.1:3000/
}
mark_current() {
  if [ -L "$root/current" ]; then
    readlink "$root/current" > "$root/previous-release"
  fi
  ln -sfn "$root/releases/$release" "$root/current"
}
previous=""
if [ -L "$root/current" ]; then previous=$(basename "$(readlink "$root/current")"); fi

if [ "$action" = deploy ]; then
  test -f "$archive" || { echo "Архив исходников не найден" >&2; exit 1; }
  # Повторный запуск того же SHA использует уже подготовленный неизменяемый каталог.
  if [ ! -d "$root/releases/$release" ]; then
    staging=$(mktemp -d "$root/releases/.staging.XXXXXX")
    trap 'rm -rf "$staging"' EXIT HUP INT TERM
    tar -xzf "$archive" -C "$staging"
    test -f "$staging/docker-compose.production.yml"
    mv "$staging" "$root/releases/$release"
    trap - EXIT HUP INT TERM
  fi
  select_release
  compose config --quiet
  compose build backend ml frontend
  compose up -d --wait --wait-timeout 120 postgres
  # Миграции выполняются до замены приложения; их совместимость проверяется перед релизом.
  compose run --rm migrate
else
  select_release
  # Откат использует сохранённые образы; пересборка и обратные миграции не выполняются.
  compose config --quiet
fi

if compose up -d --no-build --wait --wait-timeout 300 postgres ml ml-chat backend frontend && health; then
  mark_current
  echo "Релиз $release готов ($action)"
else
  echo "Проверка релиза $release не пройдена" >&2
  if [ -n "$previous" ] && [ "$previous" != "$release" ]; then
    release=$previous
    select_release
    if compose up -d --no-build --wait --wait-timeout 300 postgres ml ml-chat backend frontend && health; then
      echo "Восстановлен предыдущий релиз $release" >&2
    else
      echo "Восстановление не удалось; требуется проверка сервера" >&2
    fi
  fi
  exit 1
fi
