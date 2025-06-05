# API Документация

## Обзор API

Сервис предоставляет RESTful API для подсчета кликов по баннерам и получения статистики.

**Base URL**: `http://localhost:3000`

**Content-Type**: `application/json`

## Endpoints

### 1. Регистрация клика по баннеру

Регистрирует клик по указанному баннеру и увеличивает счетчик на +1.

**Endpoint**: `GET /counter/{bannerID}`

**Метод**: `GET`

**Параметры пути**:
- `bannerID` (integer, required) - Уникальный идентификатор баннера

**Пример запроса**:
```bash
curl -X GET http://localhost:3000/counter/1
```

**Успешный ответ** (HTTP 200):
```json
{
  "success": true,
  "message": "Click registered successfully",
  "data": {
    "bannerID": 1,
    "timestamp": "2024-12-12T10:00:10Z"
  }
}
```

**Ошибки**:

- **400 Bad Request** - Некорректный ID баннера
```json
{
  "success": false,
  "error": "Invalid banner ID",
  "code": "INVALID_BANNER_ID"
}
```

- **404 Not Found** - Баннер не найден
```json
{
  "success": false,
  "error": "Banner not found",
  "code": "BANNER_NOT_FOUND"
}
```

- **500 Internal Server Error** - Внутренняя ошибка сервера
```json
{
  "success": false,
  "error": "Internal server error",
  "code": "INTERNAL_ERROR"
}
```

### 2. Получение статистики кликов

Возвращает поминутную статистику кликов по баннеру за указанный период времени.

**Endpoint**: `POST /stats/{bannerID}`

**Метод**: `POST`

**Параметры пути**:
- `bannerID` (integer, required) - Уникальный идентификатор баннера

**Тело запроса**:
```json
{
  "from": "2024-12-12T10:00:00Z",
  "to": "2024-12-12T10:05:00Z"
}
```

**Параметры тела запроса**:
- `from` (string, required) - Начальная временная метка в формате RFC3339
- `to` (string, required) - Конечная временная метка в формате RFC3339

**Пример запроса**:
```bash
curl -X POST http://localhost:3000/stats/1 \
  -H "Content-Type: application/json" \
  -d '{
    "from": "2024-12-12T10:00:00Z",
    "to": "2024-12-12T10:05:00Z"
  }'
```

**Успешный ответ** (HTTP 200):
```json
{
  "success": true,
  "data": {
    "bannerID": 1,
    "period": {
      "from": "2024-12-12T10:00:00Z",
      "to": "2024-12-12T10:05:00Z"
    },
    "stats": [
      {
        "ts": "2024-12-12T10:00:00Z",
        "v": 4
      },
      {
        "ts": "2024-12-12T10:01:00Z",
        "v": 2
      },
      {
        "ts": "2024-12-12T10:03:00Z",
        "v": 1
      },
      {
        "ts": "2024-12-12T10:04:00Z",
        "v": 1
      }
    ],
    "total": 8
  }
}
```

**Поля ответа**:
- `bannerID` - ID баннера
- `period` - Запрошенный период
- `stats` - Массив поминутной статистики
  - `ts` - Временная метка начала минуты
  - `v` - Количество кликов в эту минуту
- `total` - Общее количество кликов за период

**Ошибки**:

- **400 Bad Request** - Некорректные параметры запроса
```json
{
  "success": false,
  "error": "Invalid request parameters",
  "code": "INVALID_PARAMETERS",
  "details": {
    "from": "Invalid timestamp format",
    "to": "End time must be after start time"
  }
}
```

- **404 Not Found** - Баннер не найден
```json
{
  "success": false,
  "error": "Banner not found",
  "code": "BANNER_NOT_FOUND"
}
```

- **422 Unprocessable Entity** - Слишком большой период запроса
```json
{
  "success": false,
  "error": "Time range too large",
  "code": "TIME_RANGE_TOO_LARGE",
  "details": {
    "maxDays": 30,
    "requestedDays": 45
  }
}
```

### 3. Получение списка баннеров (дополнительный endpoint)

Возвращает список всех доступных баннеров.

**Endpoint**: `GET /banners`

**Метод**: `GET`

**Пример запроса**:
```bash
curl -X GET http://localhost:3000/banners
```

**Успешный ответ** (HTTP 200):
```json
{
  "success": true,
  "data": {
    "banners": [
      {
        "id": 1,
        "name": "Banner 1",
        "createdAt": "2024-12-01T00:00:00Z"
      },
      {
        "id": 2,
        "name": "Banner 2",
        "createdAt": "2024-12-01T00:00:00Z"
      }
    ],
    "total": 2
  }
}
```

### 4. Health Check

Проверяет состояние сервиса и его зависимостей.

**Endpoint**: `GET /health`

**Метод**: `GET`

**Пример запроса**:
```bash
curl -X GET http://localhost:3000/health
```

**Успешный ответ** (HTTP 200):
```json
{
  "status": "healthy",
  "timestamp": "2024-12-12T10:00:00Z",
  "services": {
    "database": "healthy",
    "cache": "healthy"
  },
  "version": "1.0.0"
}
```

**Ответ при проблемах** (HTTP 503):
```json
{
  "status": "unhealthy",
  "timestamp": "2024-12-12T10:00:00Z",
  "services": {
    "database": "unhealthy",
    "cache": "healthy"
  },
  "version": "1.0.0"
}
```

## Форматы данных

### Временные метки

Все временные метки используют формат **RFC3339** (ISO 8601):
- `2024-12-12T10:00:00Z` - UTC время
- `2024-12-12T10:00:00+03:00` - Время с часовым поясом

### Коды ошибок

| Код | Описание |
|-----|----------|
| `INVALID_BANNER_ID` | Некорректный ID баннера |
| `BANNER_NOT_FOUND` | Баннер не найден |
| `INVALID_PARAMETERS` | Некорректные параметры запроса |
| `INVALID_TIMESTAMP` | Некорректный формат временной метки |
| `TIME_RANGE_TOO_LARGE` | Слишком большой временной диапазон |
| `RATE_LIMIT_EXCEEDED` | Превышен лимит запросов |
| `INTERNAL_ERROR` | Внутренняя ошибка сервера |

## Rate Limiting

Для защиты от чрезмерной нагрузки применяются следующие ограничения:

- **Counter endpoint**: 5000 запросов в секунду на IP
- **Stats endpoint**: 100 запросов в минуту на IP
- **Banners endpoint**: 60 запросов в минуту на IP

При превышении лимитов возвращается ошибка **429 Too Many Requests**:

```json
{
  "success": false,
  "error": "Rate limit exceeded",
  "code": "RATE_LIMIT_EXCEEDED",
  "retryAfter": 60
}
```

## Примеры использования

### Сценарий 1: Регистрация кликов и получение статистики

```bash
# Регистрируем несколько кликов
curl -X GET http://localhost:3000/counter/1
curl -X GET http://localhost:3000/counter/1
curl -X GET http://localhost:3000/counter/2

# Получаем статистику за последний час
curl -X POST http://localhost:3000/stats/1 \
  -H "Content-Type: application/json" \
  -d '{
    "from": "2024-12-12T09:00:00Z",
    "to": "2024-12-12T10:00:00Z"
  }'
```

### Сценарий 2: Мониторинг состояния сервиса

```bash
# Проверяем здоровье сервиса
curl -X GET http://localhost:3000/health

# Получаем список доступных баннеров
curl -X GET http://localhost:3000/banners
```

## Swagger документация

Интерактивная документация API доступна по адресу:
- **Swagger UI**: `http://localhost:3000/swagger/index.html`
- **OpenAPI JSON**: `http://localhost:3000/swagger/doc.json`

## Версионирование API

Текущая версия API: **v1**

В будущем планируется поддержка версионирования через заголовки:
```
Accept: application/vnd.api+json;version=1
```

Или через URL:
```
/api/v1/counter/{bannerID}
``` 