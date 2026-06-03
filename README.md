# Subscriptions API

REST-сервис для агрегации данных об онлайн-подписках пользователей.  
Реализован на Go с использованием PostgreSQL, миграций через goose, Docker Compose.

## Технологии

- **Go 1.26+**
- **PostgreSQL 15**
- **pgxpool** – пул соединений
- **goose** – миграции
- **chi** – роутер
- **slog** – логирование
- **Swagger** – документация API

### Предварительные требования

- Установленные **Docker** и **Docker Compose**
- Порт `8090` и `5432` свободен

### Запуск

1. Клонировать репозиторий и перейти в папку проекта:

   ```bash
   git clone https://github.com/your-username/your-repo.git
   cd your-repo
   ```

2. Создать файл `.env` в корне, пример содержания:

   ```env
   DB_HOST=postgres
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=pg_sub
   DB_PORT=5432

   POSTGRES_USER=postgres
   POSTGRES_PASSWORD=postgres
   POSTGRES_DB=pg_sub
   ```

3. Запустить сервис:

   ```bash
   docker-compose up --build
   ```

   PostgreSQL и API поднимутся, миграции применятся автоматически.

4. API будет доступен по адресу: **http://localhost:8090**

### Проверка работоспособности

  Запуск тестов с помощью testify/httptest (требуется на основной машине): `go test ./test -v`, перед запуском команды необходимо поднять БД и добавить в директорию тестов `.env` файл с параметрами подключения к ней.
  
## Документация API (Swagger)

После запуска сервера перейти в браузере:  
**http://localhost:8090/swagger/index.html**

Откроется описание всех эндпоинтов, их параметры и примеры ответов.
