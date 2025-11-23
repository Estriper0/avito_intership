# PR Reviewer Assignment Service

Микросервис, который автоматически назначает ревьюеров на Pull Request’ы (PR), а также позволяет управлять командами и участниками. Взаимодействие происходит исключительно через HTTP API. Проект следует чистой архитектуре (Clean Architecture). Слои независимы: внешние зависимости легко заменяемы.

## Технологии

- **Go 1.21+**
- **PostgreSQL** — основное хранилище пользователей
- **Docker & Docker Compose** - для запуска приложения
- **Gin** - веб фреймворк
- **pgx** - для работы с PostgreSQL
- **avito-tech/go-transaction-manager** - менеджер транзакций
- **go-playground/validator** - для валидации входных данных
- **golang-migrate/migrate** - миграции базы данных
- **testcontainers/testcontainers-go** - для интеграционных тестов

---

# Шаги по запуску

1. **Клонируй репозиторий и перейдите в папку**:
   ```
   git clone https://github.com/Estriper0/avito_intership.git
   cd avito_intership
   ```
2. Создайте файл `.env` и настройте переменные окружения:
   ```env
    ENV=local

    DB_HOST=postgres
    DB_PORT=5432
    DB_USER=postgres
    DB_PASSWORD=12345
    DB_NAME=postgres

    DB_URL=postgres://postgres:12345@postgres:5432/postgres
   ```

3. **Запусти с помощью Make**:
   ```
   make up
   ```

---

## Тестирование

Для запуска тестов выполните:
```bash
make test
```

---

## Реализованные endpoints:
### Основные по заданию:
    1. /team/add - Создание команды
    2. /team/get - Получение команды
    3. /users/setIsActive - Активация/деактивация пользоавателя
    4. /users/getReview - Получение всех пул реквестов, где пользователь назначен ревьюером
    5. /pullRequest/create - Создание пул реквеста с назначение 2 случайных ревьюеров из команды автора
    6. /pullRequest/merge - Операция merge (меняем статус)
    7. /pullRequest/reassign - Переназначаем ревьюера на другого случайного из команды пользователя
### Дополнительно реализовано:
1. /team/stats/pull_request - Получаем статистику пул реквестов по командам
Пример запроса:
```bash
/team/stats/pull_request?team_name=payments
```

Пример ответа:
```json
{
    "team_name": "payments",
    "total_pull_request": 12,
    "open_pull_request": 7,
    "merged_pull_request": 5
}
```
2. /users/stats/review - Получаем статистику назначенных открытых пул реквестов по пользователям
Пример ответа:
```json
{
"users": [
        {
            "user_id": "u045",
            "username": "Samuel",
            "count_open_review": 4
        },
        {
            "user_id": "u068",
            "username": "Piper",
            "count_open_review": 5
        }
    ]
}
```
3. /users/massDeactivation - Массовая деактивация пользователей (в запросе нужен хотя бы один существующий пользователь, иначе 404)
Пример запроса:
{
    "users_id": ["u001", "u002"]
}
Пример ответа:
```json
{
    "deactivated_users_id": [
        "u001",
        "u002"
    ]
}
```
4. /pullRequest/reassign/team - Переназначем всех неактивных ревьюеров в команде. Если подходящих кандидатов нет, то оставляем как есть.
Пример запроса:
```bash
/pullRequest/reassign/team?team_name=payments
```

Пример ответа:
```json
{
    "message": "The task has been received"
}
```
Задача ставится в очередь и обрабатывается. В качестве очереди используются каналы (быстрая реализация без RabbitMQ/Kafka). Несколько воркеров читают канал и запускают обработку.