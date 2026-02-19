# Task-Manager

A scalable, secure RESTful Task Management Service built with **Go** (Golang 1.24.0), PostgreSQL, JWT authentication, and background task processing.

## Features

- **User Authentication & Authorization**
  - Register / Login with JWT tokens
  - Role-based access: normal users see only their tasks, admins see all tasks
- **Task CRUD Operations**
  - Create, Read (list & single), Update, Delete tasks
  - Tasks are owned by users (enforced via authorization)
- **Background Auto-Completion**
  - Tasks in `pending` or `in_progress` are automatically marked `completed` after configurable delay
  - Uses goroutines + channel-based worker queue
  - Safe: skips already completed or deleted tasks
- **Pagination & Filtering**
  - List tasks with `?limit=`, `?offset=`, `?status=pending|in_progress|completed`
- **Rate Limiting**
  - Configurable per-minute request limit per IP
- **Swagger Documentation**
  - Interactive API docs at `/swagger/index.html`
- **Graceful Shutdown**
  - Handles SIGTERM/SIGINT cleanly
- **Environment-based Configuration**
  - Uses `.env` file via viper

## Tech Stack

- **Language**: Go 1.24.0
- **Framework/Router**: Gorilla Mux
- **Database**: PostgreSQL (with `lib/pq`)
- **Authentication**: JWT (HS256)
- **Password Hashing**: bcrypt
- **ORM/Access Layer**: Plain SQL + Repository pattern
- **Concurrency**: Goroutines + buffered channels
- **Middleware**: Authentication, Rate Limiting
- **Documentation**: Swagger (swaggo)
- **CORS & Negroni** for middleware chaining

## Project Structure
