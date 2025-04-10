Вот пример README для вашего проекта на двух языках (русский и английский):

```markdown
# Прокси-сервер с динамическим реверс-прокси и поддержкой TCP/UDP

Этот проект представляет собой прокси-сервер на Go, который динамически читает правила переадресации из базы данных MySQL и проксирует соединения через TCP и UDP. Сервер поддерживает проксирование как для TCP, так и для UDP, а также для обоих протоколов одновременно. Настройки для подключения к базе данных считываются из конфигурационного файла в формате YAML.

## Функции:
- Прокси-сервер с поддержкой TCP и UDP.
- Динамическая синхронизация с базой данных для добавления, удаления и обновления правил.
- Конфигурация подключения к базе данных из файла `config.yaml`.
- Автоматическое добавление/удаление прослушиваемых портов в зависимости от базы данных.

## Установка

1. Установите Go:
   Скачайте и установите Go с [официального сайта](https://golang.org/dl/).

2. Клонируйте репозиторий:
   ```bash
   git clone https://github.com/Rxflex/chukproxy.git
   cd chukproxy
   ```

3. Установите зависимости:
   ```bash
   go mod tidy
   ```

4. Создайте конфигурационный файл `config.yaml`:
   Пример конфигурационного файла:
   ```yaml
   database:
     user: "your_user"
     password: "your_password"
     host: "127.0.0.1"
     port: 3306
     dbname: "your_db"
   ```

5. Запустите сервер:
   ```bash
   go run main.go
   ```

## Структура проекта

- `main.go`: Основной файл с логикой работы прокси-сервера.
- `config.yaml`: Конфигурационный файл для подключения к базе данных.

## Конфигурация базы данных

Прокси-сервер использует базу данных MySQL для получения правил переадресации. Пример SQL-запроса для создания таблицы с правилами:

```sql
CREATE TABLE routes (
    listen_port INT NOT NULL,
    target_ip VARCHAR(15) NOT NULL,
    target_port INT NOT NULL,
    protocol ENUM('tcp', 'udp', 'both') NOT NULL,
    PRIMARY KEY (listen_port)
);
```

## Лицензия

Этот проект распространяется под лицензией MIT. См. файл [LICENSE](LICENSE) для подробностей.

---

# Reverse Proxy Server with Dynamic Rule Synchronization and TCP/UDP Support

This project is a Go-based reverse proxy server that dynamically reads redirection rules from a MySQL database and proxies connections over TCP and UDP. The server supports proxying both TCP and UDP connections, as well as both protocols simultaneously. The database connection settings are read from a YAML configuration file.

## Features:
- Reverse proxy server with TCP and UDP support.
- Dynamic synchronization with a MySQL database to add, remove, and update rules.
- Database connection configuration from the `config.yaml` file.
- Automatic addition/removal of listening ports based on database entries.

## Installation

1. Install Go:
   Download and install Go from the [official website](https://golang.org/dl/).

2. Clone the repository:
   ```bash
   git clone https://github.com/Rxflex/chukproxy.git
   cd chukproxy
   ```

3. Install dependencies:
   ```bash
   go mod tidy
   ```

4. Create the `config.yaml` configuration file:
   Example configuration file:
   ```yaml
   database:
     user: "your_user"
     password: "your_password"
     host: "127.0.0.1"
     port: 3306
     dbname: "your_db"
   ```

5. Run the server:
   ```bash
   go run main.go
   ```

## Project Structure

- `main.go`: Main file containing the reverse proxy server logic.
- `config.yaml`: Configuration file for the database connection.

## Database Configuration

The proxy server uses a MySQL database to retrieve redirection rules. Example SQL query to create a table with rules:

```sql
CREATE TABLE routes (
    listen_port INT NOT NULL,
    target_ip VARCHAR(15) NOT NULL,
    target_port INT NOT NULL,
    protocol ENUM('tcp', 'udp', 'both') NOT NULL,
    PRIMARY KEY (listen_port)
);
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.