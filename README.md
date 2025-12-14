# 🏆 Real-Time Leaderboard System

<div align="center">

![Go](https://img.shields.io/badge/go-%2300ADD8.svg?style=for-the-badge&logo=go&logoColor=white)
![Postgres](https://img.shields.io/badge/postgres-%23316192.svg?style=for-the-badge&logo=postgresql&logoColor=white)
![Redis](https://img.shields.io/badge/redis-%23DD0031.svg?style=for-the-badge&logo=redis&logoColor=white)
![RabbitMQ](https://img.shields.io/badge/Rabbitmq-FF6600?style=for-the-badge&logo=rabbitmq&logoColor=white)
![JWT](https://img.shields.io/badge/JWT-black?style=for-the-badge&logo=JSON%20web%20tokens)
![Docker](https://img.shields.io/badge/docker-%230db7ed.svg?style=for-the-badge&logo=docker&logoColor=white)

*Распределённая backend система для отслеживания результатов игроков в реальном времени*

</div>

## 🌟 Обзор

Real-Time Leaderboard — это высокопроизводительная распределённая система для управления игровыми таблицами лидеров. Система обеспечивает мгновенное обновление рейтингов, JWT-аутентификацию и верификацию по email с использованием микросервисной архитектуры.

## 🏗️ Архитектура

Система состоит из 6 компонентов:

```
┌─────────────┐
│   Клиент    │
└──────┬──────┘
       │
       ▼
┌──────────────────────────────────────────────┐
│         Leaderboard Service                  │
│  (API для управления таблицей лидеров)       │
└────┬─────────────────────────────────┬───────┘
     │                                  │
     │                                  │
     ▼                                  ▼
┌────────────┐                   ┌─────────────┐
│   Redis    │                   │Auth Service │
│(Leaderboard│                   │   (JWT +    │
│   Data)    │                   │Registration)│
└────────────┘                   └──────┬──────┘
                                        │
                                        ▼
                                 ┌──────────────┐
                                 │  RabbitMQ    │
                                 │  (Email      │
                                 │   Queue)     │
                                 └──────┬───────┘
                                        │
                    ┌───────────────────┴──────────┐
                    │                              │
                    ▼                              ▼
            ┌──────────────┐              ┌───────────────┐
            │ PostgreSQL   │              │Email Sender   │
            │(User Data)   │              │   Service     │
            └──────────────┘              └───────────────┘
```

### Компоненты

1. **Auth Service** - Сервис аутентификации и авторизации пользователей через JWT и Refresh токены
2. **Email Sender Service** - Асинхронная отправка писем для подтверждения email при регистрации
3. **Leaderboard Service** - Основной API сервис для управления таблицей лидеров и отправки результатов
4. **RabbitMQ** - Брокер сообщений для передачи задач на отправку email
5. **PostgreSQL** - Хранит данные о пользователях (учётные записи, credentials)
6. **Redis** - In-memory хранилище для таблиц лидеров с использованием Sorted Sets и Lua скриптов

## ✨ Возможности

- 🚀 **Обновления в реальном времени** - Мгновенное обновление рейтингов с использованием Redis
- 🔐 **JWT Аутентификация** - Защищённая авторизация с Access и Refresh токенами
- 📧 **Email Верификация** - Подтверждение email при регистрации
- 🎮 **Мультиигровая поддержка** - Отдельные таблицы лидеров для разных игр
- 📊 **Быстрые запросы** - Redis Sorted Sets для O(log N) операций с рейтингами
- ⚡ **Lua Скрипты** - Атомарные операции на стороне Redis для производительности
- 🔄 **Асинхронная обработка** - RabbitMQ для отправки email без блокировки основных операций
- 🐳 **Контейнеризация** - Полная поддержка Docker для всех сервисов

## 🛠️ Технологический стек

- **Backend**: Go
- **Аутентификация**: JWT (Access + Refresh tokens)
- **База данных**: PostgreSQL (пользователи)
- **Кэш/Хранилище**: Redis (таблицы лидеров) + Lua скрипты
- **Message Broker**: RabbitMQ
- **Контейнеризация**: Docker & Docker Compose

## 📋 Требования

- Go 1.21+
- Docker & Docker Compose
- PostgreSQL 15+
- Redis 7.0+
- RabbitMQ 3.12+

## 📡 API Документация

### Auth Service

#### Регистрация пользователя

```http
POST /register
Content-Type: application/json

{
  "email": "user@example.com",
  "username": "player1",
  "password": "securepass123"
}
```

**Ответ:**
```json
{
  "user_id": 123
}
```

#### Вход в систему

```http
POST /login
Content-Type: application/json

{
  "email": "user@example.com",
  "password": "securepass123",
  "app_id": 1
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### Выход из системы

```http
POST /logout
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### Обновление токена

```http
POST /refresh
Content-Type: application/json

{
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

**Ответ:**
```json
{
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs..."
}
```

#### Подтверждение email

```http
GET /verify?token=verification_token
```

### Leaderboard Service

**Все endpoint'ы требуют JWT токен в заголовке:**
```http
Authorization: Bearer <access_token>
```

#### Отправить результат

```http
POST /submit
Content-Type: application/json
Authorization: Bearer <access_token>

{
  "game": "snake",
  "score": 4200
}
```

**Ответ:**
```json
{
  "game": "snake",
  "score": 4200,
  "rank": 5
}
```

#### Получить топ игроков

```http
GET /{game}/top?limit=20
Authorization: Bearer <access_token>
```

**Ответ:**
```json
{
  "game": "snake",
  "top": [
    {
      "user_id": 1,
      "username": "u1",
      "score": 9000,
      "rank": 1
    },
    {
      "user_id": 2,
      "username": "u2",
      "score": 8000,
      "rank": 2
    }
  ]
}
```

#### Получить свой результат

```http
GET /{game}/me
Authorization: Bearer <access_token>
```

**Ответ:**
```json
{
  "game": "snake",
  "user_id": 123,
  "score": 4200,
  "rank": 5
}
```

## 📊 Особенности системного дизайна

### JWT Authentication Flow

1. Пользователь регистрируется → получает `user_id`
2. Email отправляется в RabbitMQ для асинхронной отправки
3. Пользователь подтверждает email по ссылке
4. Пользователь входит → получает `access_token` и `refresh_token`
5. При истечении `access_token` используется `refresh_token` для получения новой пары
6. Все запросы к Leaderboard Service требуют валидный `access_token`

### Email Verification Flow

1. При регистрации создаётся verification token
2. Задача отправляется в RabbitMQ очередь
3. Email Sender Service асинхронно обрабатывает очередь
4. Отправляется письмо с ссылкой подтверждения
5. Пользователь переходит по ссылке `/verify?token=...`
6. Email подтверждается, учётная запись активируется

### Преимущества архитектуры

- **Разделение ответственности** - каждый сервис решает одну задачу
- **Горизонтальная масштабируемость** - можно запустить несколько инстансов каждого сервиса
- **Fault Tolerance** - отказ одного сервиса не ломает всю систему
- **Асинхронность** - отправка email не блокирует регистрацию
- **Производительность** - Redis обеспечивает sub-millisecond latency

Этот проект создан на основе задания от [roadmap.sh/projects/realtime-leaderboard-system](https://roadmap.sh/projects/realtime-leaderboard-system)

</div>
