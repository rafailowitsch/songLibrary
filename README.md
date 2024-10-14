### Описание

Приложение представляет собой реализацию онлайн-библиотеки песен. Оно поддерживает следующие REST методы:

- Получение данных библиотеки с фильтрацией по всем полям и пагинацией.
- Получение текста песни с пагинацией по куплетам.
- Удаление песни.
- Изменение данных песни.
- Добавление новой песни.

**Технологический стек:** Go (Golang), Chi, PostgreSQL, Redis, Docker.

В корневой папке находится директория `mockMusicInfo`, которая содержит моковое стороннее API, с которым взаимодействует основное приложение при добавлении песни. В 30% случаев это API возвращает ошибку `BadRequest`, чтобы симулировать отсутствие песни в базе данных. Это API запускается командой:

```sh
go run mockMusicInfo/mockMusicInfo.go
```

### Настройка

Для запуска приложения необходимо создать `.env` файл со следующими переменными:

```plaintext
POSTGRES_PASSWORD=пароль для PostgreSQL
REDIS_PASSWORD=пароль для Redis
CONFIG_PATH=путь до конфигурационного файла
```

После настройки конфигурации Swagger будет доступен по адресу: [http://localhost:8089/swagger/](http://localhost:8089/swagger/).

### Изменение уровня логирования

Чтобы изменить уровень логирования в приложении, необходимо изменить переменную `env` в файле `./config/config.yaml`. Приложение поддерживает три уровня логирования:

- **local**: Логи выводятся в удобном для чтения формате с уровнями `info`, `warn` и `debug`.
- **dev**: Логи выводятся в формате JSON с уровнями `info`, `warn` и `debug`.
- **prod**: Логи выводятся в формате JSON с уровнями `info` и `warn` (уровень `debug` отключен).

Пример изменения уровня логирования в файле `config.yaml`:

```yaml
env: local  # Возможные значения: local, dev, prod
```

Измените значение переменной `env` на нужный уровень в зависимости от того, как вы планируете использовать приложение.

### Миграции

Для применения или отката миграций воспользуйтесь следующими командами (таблица `songs` создаётся автоматически при запуске приложения через миграции):

- **Применение миграции:**

```sh
migrate -path ./internal/app/migrations -database 'postgres://postgres:postgres@localhost:5434/postgres?sslmode=disable' up 1
```

- **Откат миграции:**

```sh
migrate -path ./internal/app/migrations -database 'postgres://postgres:postgres@localhost:5434/postgres?sslmode=disable' down 1
```

## Запуск и использование приложения

Выполните следующие команды для сборки и запуска приложения:

```sh
docker-compose up --build -d
go run cmd/main.go
```

### Примеры использования API

#### POST: /songs

Добавляет новую песню в библиотеку.

**Пример запроса:**

```sh
curl -X POST localhost:8089/songs/ -H "Content-Type: application/json" -d '{
    "name": "Mr. Blue Sky",
    "group": "ELO"
}'
```

#### GET: /songs

Получает список всех песен с возможностью фильтрации по параметрам.

**Пример запроса:**

```sh
curl -X GET localhost:8089/songs?group=elo -H "Content-Type: application/json"
```

**Пример ответа:**

```json
[
    {
        "id": "51ee20ca-35a3-4da6-9111-b796b56adfb2",
        "name": "Mr. Blue Sky",
        "group": "ELO",
        "text": "Ooh baby, can you hear me moan?",
        "link": "https://example.com/song4",
        "release_date": "2024-10-14T00:00:00Z",
        "created_at": "2024-10-14T23:36:29.170294Z",
        "updated_at": "2024-10-14T23:36:29.170294Z"
    },
    {
        "id": "fe88a8db-fbd0-47e4-805c-c1293f5b79b8",
        "name": "Evil Woman",
        "group": "ELO",
        "text": "Ooh baby, don't you know I suffer?",
        "link": "https://example.com/song3",
        "release_date": "2024-10-14T00:00:00Z",
        "created_at": "2024-10-14T23:36:46.418175Z",
        "updated_at": "2024-10-14T23:36:46.418175Z"
    },
    {
        "id": "52f26279-c8e1-4e71-9482-a7e105c3d35a",
        "name": "Evil Woman",
        "group": "ELO",
        "text": "You caught me under false pretenses.",
        "link": "https://example.com/song4",
        "release_date": "2024-10-14T00:00:00Z",
        "created_at": "2024-10-14T23:37:40.34581Z",
        "updated_at": "2024-10-14T23:37:40.34581Z"
    }
]
```
