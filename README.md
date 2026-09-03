# Currency Exchange

Асинхронный сервис котировок валют. Хранит курсы в PostgreSQL, обновляет их по запросу через внешний API ([CurrencyBeacon](https://currencybeacon.com/)), отдаёт REST API и Swagger UI.

## Стек

- Go
- PostgreSQL 16
- chi (HTTP router)
- gocron (scheduler)
- swag / http-swagger
- Docker Compose

## Возможности

- Инициализация стартовых пар валют при запуске
- Получение текущего курса
- Асинхронный запрос на обновление курса
- Получение результата обновления по `request id`
- История изменений курса
- Фоновый scheduler, который обрабатывает незавершённые update-request'ы

Формат имени пары: `EUR_USD`, `USD_MXN` (3 заглавные буквы + `_` + 3 заглавные буквы).

## Быстрый старт (Docker)

1. Скопируй env:

```bash
cp .env.example .env
```

2. Укажи ключ API в `.env`:

```env
API_KEY=your_currencybeacon_api_key
```

3. Запусти:

```bash
docker compose up --build
```

Сервис будет доступен на:

- API: http://localhost:8090
- Swagger: http://localhost:8090/swagger/index.html
- PostgreSQL: `localhost:5432` (`postgres` / `postgres`, БД `currency_exchange`)

## API

| Method | Path | Описание |
|--------|------|----------|
| `GET` | `/quotation/{name}` | Текущий курс |
| `POST` | `/quotation/{name}/update` | Создать async-запрос на обновление |
| `GET` | `/quotation/request/{update_id}` | Результат обновления по id запроса |
| `GET` | `/quotation/{name}/updates` | История обновлений |

### Примеры

Получить курс:

```bash
curl http://localhost:8090/quotation/EUR_USD
```

Запросить обновление:

```bash
curl -X POST http://localhost:8090/quotation/EUR_USD/update \
  -H 'Content-Type: application/json' \
  -d '{"quotation_name":"EUR_USD"}'
```

Ответ:

```json
{
  "status": "OK",
  "data": { "quotation_request_id": 1 },
  "error": ""
}
```

Проверить статус запроса:

```bash
curl http://localhost:8090/quotation/request/1
```

Пока scheduler ещё не обработал запрос, вернётся `409` (`quotation request not done`). После обработки — актуальный курс.

История обновлений:

```bash
curl http://localhost:8090/quotation/EUR_USD/updates
```

## Структура проекта

```text
cmd/currency-exchange/     # точка входа
config/                    # yaml-конфиги
docs/                      # сгенерированный swagger
internal/
  config/                  # загрузка конфигурации
  http-server/             # handlers + middleware
  models/                  # доменные модели
  service/
    quotation/             # бизнес-логика
    quotationAPI/          # клиент внешнего API
  sheduler/                # фоновый scheduler
  storage/                 # интерфейс + postgres
  test/mock/               # моки для unit-тестов
```

## Как работает обновление

1. `POST /quotation/{name}/update` создаёт запись в `quotation_request` (`done=false`) и сразу возвращает `id`.
2. Scheduler периодически забирает незавершённые запросы.
3. Для каждого запроса сервис ходит во внешний API, обновляет `quotation`, пишет историю в `quotation_update`, помечает request как `done`.
4. `GET /quotation/request/{id}` до завершения отдаёт ошибку, после — актуальный курс.
