# Click Counter Service

Высокопроизводительный сервис счетчика кликов.

## 🚀 Быстрый старт

### Предварительные требования
- Docker 20.10+
- Docker Compose 2.0+

####Load test
docker/loadtest/results.md

## 🎯 Результаты тестирования

| Target RPS | Actual RPS | Success Rate | Avg Latency | P95 Latency | Workers | Total Requests | Failed | Status |
|------------|------------|--------------|-------------|-------------|---------|----------------|--------|---------|
| **100**    | **99.0**   | **100.0%** ✅ | 956µs       | 1.441ms     | 10      | 1,990          | 0      | ✅ Отлично |
| **200**    | **198.0**  | **100.0%** ✅ | 1.063ms     | 1.361ms     | 20      | 3,976          | 0      | ✅ Отлично |
| **300**    | **296.9**  | **100.0%** ✅ | 1.418ms     | 1.663ms     | 30      | 5,968          | 0      | ✅ Отлично |
| **400**    | **395.9**  | **100.0%** ✅ | 1.925ms     | 2.282ms     | 40      | 7,958          | 0      | ✅ Отлично |
| **500**    | **494.7**  | **97.3%** ⚠️  | 2.380ms     | 3.234ms     | 50      | 9,947          | 268    | ⚠️ Деградация |
| **600**    | **593.9**  | **79.3%** ❌  | 2.432ms     | 4.455ms     | 60      | 11,936         | 2,476  | ❌ Критично |


### Запуск системы

```bash
cd docker
docker-compose up -d
```

Система будет доступна по адресу: http://localhost:8080

### Остановка системы

```bash
cd docker
docker-compose down
```

## 📋 API Endpoints

| Endpoint | Метод | Описание |
|----------|-------|----------|
| `/health` | GET | Проверка здоровья сервиса |
| `/counter/{id}` | GET | Регистрация клика по баннеру |
| `/stats/{id}` | POST | Получение статистики кликов |

### Примеры использования

```bash
# 1. Проверить здоровье сервиса
curl http://localhost:8080/health


# 2. Зарегистрировать клик по баннеру 1
curl http://localhost:8080/counter/1

# 3. Получить статистику за период
curl -X POST http://localhost:8080/stats/1 \
  -H "Content-Type: application/json" \
  -d '{
    "from": "2024-01-01T00:00:00Z",
    "to": "2025-01-01T00:00:00Z"
  }'
```

## 🏗️ Архитектура проекта

```
clickcounter/
├── app/                    # Основное Go приложение
│   ├── cmd/               # Точки входа
│   ├── internal/          # Внутренняя логика
│   ├── configs/           # Конфигурационные файлы
│   ├── migrations/        # Миграции базы данных
│   └── docs/              # Swagger документация
├── docker/                # Docker окружение
│   ├── docker-compose.yml # Конфигурация сервисов
│   └── Dockerfile         # Образ приложения
└── README.md             # Этот файл
```

### Архитектура системы

```
┌─────────────────┐    ┌─────────────────┐
│   ClickCounter  │    │   PostgreSQL    │
│   Application   │────│    Database     │
│   Port: 8080    │    │   Port: 5432    │
└─────────────────┘    └─────────────────┘
```

### Компоненты

- **PostgreSQL 15**: Основная база данных с оптимизированными индексами
- **ClickCounter App**: Go приложение с memory кэшем и batch обработкой
- **Migrate**: Автоматические миграции базы данных

## ⚙️ Управление системой

### Основные команды

```bash
cd docker

# Запуск системы
docker-compose up -d

# Остановка системы
docker-compose down

# Перезапуск системы
docker-compose restart

# Пересборка и запуск
docker-compose up -d --build

# Просмотр статуса сервисов
docker-compose ps

# Просмотр логов всех сервисов
docker-compose logs -f

# Просмотр логов конкретного сервиса
docker-compose logs -f clickcounter
docker-compose logs -f postgres
```

### Режимы запуска

#### 1. Production (полная система)
```bash
cd docker
docker-compose up -d
# Доступ через http://localhost:8080
```

#### 2. Development (только база данных)
```bash
cd docker
# Запустить только PostgreSQL
docker-compose up -d postgres

# В другом терминале - запустить приложение локально
cd ../app
CLICKCOUNTER_DATABASE_HOST=localhost go run ./cmd/server
```

## 📊 Мониторинг и управление

### Просмотр логов
```bash
cd docker

# Все сервисы
docker-compose logs -f

# Только приложение
docker-compose logs -f clickcounter

# Только база данных
docker-compose logs -f postgres

# Последние 100 строк логов
docker-compose logs --tail=100 clickcounter
```

### Управление данными
```bash
cd docker

# Подключиться к базе данных
docker-compose exec postgres psql -U clickcounter -d clickcounter

# Создать backup базы данных
docker-compose exec postgres pg_dump -U clickcounter clickcounter > backup.sql

# Восстановить из backup
docker-compose exec -T postgres psql -U clickcounter -d clickcounter < backup.sql

# Очистить данные (удалить volumes)
docker-compose down -v
```

### Отладка и диагностика
```bash
cd docker

# Подключиться к контейнеру приложения
docker-compose exec clickcounter sh

# Проверить конфигурацию приложения
docker-compose exec clickcounter cat /app/configs/config.docker.yaml

# Проверить переменные окружения
docker-compose exec clickcounter env | grep CLICKCOUNTER

# Проверить сетевое подключение
docker-compose exec clickcounter ping postgres
```


## 🔧 Разработка

### Локальная разработка

```bash
# Запустить только базу данных
cd docker
docker-compose up -d postgres

# В другом терминале - запустить приложение
cd ../app
CLICKCOUNTER_DATABASE_HOST=localhost go run ./cmd/server
```

### Пересборка приложения

```bash
cd docker

# Пересобрать и перезапустить приложение
docker-compose up -d --build clickcounter

# Перезапустить только приложение (без пересборки)
docker-compose restart clickcounter
```

### Работа с миграциями

```bash
cd docker

# Применить миграции вручную
docker-compose run --rm migrate \
  -path /migrations \
  -database "postgres://clickcounter:clickcounter_password@postgres:5432/clickcounter?sslmode=disable" \
  up

# Откатить последнюю миграцию
docker-compose run --rm migrate \
  -path /migrations \
  -database "postgres://clickcounter:clickcounter_password@postgres:5432/clickcounter?sslmode=disable" \
  down 1
```


### Конфигурация
Основные настройки можно изменить через переменные окружения в `docker-compose.yml`:

```yaml
environment:
  # База данных
  CLICKCOUNTER_DATABASE_HOST: postgres
  CLICKCOUNTER_DATABASE_MAX_OPEN_CONNS: 200
  
  # Производительность
  CLICKCOUNTER_CLICK_FLUSHER_BATCH_SIZE: 2000
  CLICKCOUNTER_CLICK_FLUSHER_INTERVAL: 2
  
  # Логирование
  CLICKCOUNTER_LOGGER_LEVEL: info
```


## 🏆 Особенности реализации

### Архитектурные принципы
- **DDD (Domain Driven Design)**: Доменно-ориентированное проектирование

### Технологический стек
- **Go 1.23**: Основной язык программирования
- **PostgreSQL 15**: Реляционная база данных


