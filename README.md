# Fitness App (Спортик)

## Запуск (корень репозитория)

**1. Postgres**

```powershell
docker compose up -d postgres
```

**2. ML (FastAPI, порт 5050)**

```powershell
cd ai
python -m pip install -r requirements-ml-api.txt
python -m uvicorn main:app --host 127.0.0.1 --port 5050
```

**3. Backend**

```powershell
cd backend
$env:MY_KEY="postgres://sport:sport@localhost:5433/sport?sslmode=disable"
$env:ML_BASE_URL="http://localhost:5050"
go run .
```

**4. Frontend**

```powershell
cd frontend
npm install
npm run dev
```

→ **http://localhost:5173/**

---

## БД посмотреть

```powershell
docker compose exec postgres psql -U sport -d sport
```

В psql: `\dt` — таблицы, `SELECT * FROM users LIMIT 5;`, `\q` — выход.


