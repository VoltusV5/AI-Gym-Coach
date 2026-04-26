# Проверка цепочки питания: guest → POST entry → GET entries → GET dashboard.
# Один JWT на все запросы. Без этого GET вернёт пустой список (другой пользователь).
#
# Запуск из корня репозитория:
#   powershell -ExecutionPolicy Bypass -File scripts/verify-nutrition-api.ps1
# Через nginx (порт 3000):
#   $env:NUTRITION_API_BASE = "http://127.0.0.1:3000"; powershell -ExecutionPolicy Bypass -File scripts/verify-nutrition-api.ps1

$ErrorActionPreference = "Stop"
$base = if ($env:NUTRITION_API_BASE) { $env:NUTRITION_API_BASE.TrimEnd("/") } else { "http://127.0.0.1:5050" }
$day = (Get-Date).ToString("yyyy-MM-dd")

Write-Host "Base: $base | day: $day"

$token = (
  curl.exe -sS -X POST "$base/api/v1/auth/guest" -H "Content-Type: application/json" -d "{}"
) | ConvertFrom-Json | Select-Object -ExpandProperty token
if (-not $token) { throw "No token from /api/v1/auth/guest" }

$jsonBody = (@{
  meal_type = "breakfast"
  grams     = 150
  title     = "verify script"
  protein_g = 10
  fat_g     = 5
  carbs_g   = 30
  calories  = 210
  day       = $day
} | ConvertTo-Json -Compress)

$tmp = [System.IO.Path]::GetTempFileName()
try {
  [System.IO.File]::WriteAllText($tmp, $jsonBody, [System.Text.UTF8Encoding]::new($false))

  $postOut = curl.exe -sS -f -X POST "$base/api/v1/nutrition/entries" `
    -H "Authorization: Bearer $token" -H "Content-Type: application/json; charset=utf-8" `
    --data-binary "@$tmp"
  if (-not $postOut) { throw "POST entries: empty body" }
  $created = $postOut | ConvertFrom-Json
  if (-not $created.id) { throw "POST entries: no id" }

  $listRaw = curl.exe -sS "$base/api/v1/nutrition/entries?day=$day" -H "Authorization: Bearer $token"
  $list = $listRaw | ConvertFrom-Json
  $ids = @($list.items | ForEach-Object { $_.id })
  if ($ids -notcontains $created.id) {
    throw "GET entries: id $($created.id) not in [$($ids -join ',')]. Raw: $listRaw"
  }

  $dashRaw = curl.exe -sS "$base/api/v1/nutrition/dashboard?day=$day" -H "Authorization: Bearer $token"
  $dash = $dashRaw | ConvertFrom-Json
  $k = [double]$dash.today.calories
  if ($k -le 0) { throw "GET dashboard: today.calories is 0. Raw: $dashRaw" }

  Write-Host "OK: entry id=$($created.id), list ids=$($ids -join ','), dashboard today.calories=$k"
}
finally {
  Remove-Item -LiteralPath $tmp -Force -ErrorAction SilentlyContinue
}
