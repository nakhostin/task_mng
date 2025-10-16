# XDR - سیستم مدیریت Task

یک API برای مدیریت Task توسعه داده شده با Go که قابلیت‌های مدیریت کامل Task، احراز هویت با JWT، کش با Redis و معماری DDD را پیاده‌سازی کرده است.

## قابلیت‌های سیستم

### قابلیت‌های کاربردی
- **احراز هویت کاربر**: ثبت نام، ورود به سیستم و refresh token
- **مدیریت Task**: عملیات CRUD کامل برای وظایف
- **تخصیص Task**: امکان اختصاص وظایف به کاربران مختلف
- **تغییر وضعیت**: تغییر وضعیت وظایف از ToDo به InProgress و Done
- **فیلتر پیشرفته**: فیلتر بر اساس assignee، status و priority
- **Pagination**: صفحه‌بندی برای مدیریت داده‌های حجیم

### ویژگی‌های فنی
- کش‌گذاری با Redis برای بهینه‌سازی عملکرد
- احراز هویت مبتنی بر JWT
- مستندات خودکار با Swagger
- پشتیبانی کامل از Docker
- تست‌های واحد
- معماری Clean Architecture

## ساختار کلی پروژه

این پروژه از معماری DDD استفاده می‌کند تا کد تمیز و قابل نگهداری باشد:

```
┌─────────────────────────────────────────┐
│         Presentation Layer              │
│    (HTTP Handlers, Middleware)          │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Application Layer               │
│          (Services)                     │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│         Domain Layer                    │
│    (Entities, Repository)               │
└──────────────┬──────────────────────────┘
               │
┌──────────────▼──────────────────────────┐
│      Infrastructure Layer               │
│    (Database, Redis, JWT)               │
└─────────────────────────────────────────┘
```

## نحوه اجرا

### با Docker (پیشنهادی)

#### مرحله اول: Build کردن Image

```bash
docker build -t xdr:latest -f cmd/web/Dockerfile .
```

#### مرحله دوم: اجرای سرویس‌ها

```bash
docker-compose up -d
```

این دستور سرویس‌های زیر را اجرا می‌کند:
- **API Server**: روی پورت `8088`
- **PostgreSQL**: روی پورت `5432`
- **Redis**: روی پورت `6379`
- **Prometheus**: روی پورت `9090`
- **Grafana**: روی پورت `3000`


#### دسترسی به API

پس از اجرای موفقیت‌آمیز، می‌توانید از آدرس‌های زیر استفاده کنید:
- **آدرس API**: `http://localhost:8088/api/v1`
- **Swagger**: `http://localhost:8088/swagger/index.html`

#### توقف سرویس‌ها

```bash
docker-compose down
```

برای حذف volumes و داده‌ها:

```bash
docker-compose down -v
```

### اجرای محلی (بدون Docker)

در صورت تمایل به اجرای پروژه بدون Docker:

#### نصب Dependencies

```bash
go mod download
```

#### نصب Swag (برای Swagger)

```bash
go install github.com/swaggo/swag/cmd/swag@latest
export PATH=$PATH:$(go env GOPATH)/bin
```

#### ساخت Swagger

```bash
swag init -g cmd/web/main.go -o ./docs
```

#### تنظیم Config

یک فایل `.env` در ریشه پروژه ایجاد کنید:

```env
HOST=0.0.0.0
PORT=8088

JWT_ACCESS_SECRET=your-secret-key-change-this
JWT_REFRESH_SECRET=your-refresh-secret-key
JWT_ACCESS_TTL=24h
JWT_REFRESH_TTL=720h
JWT_ISSUER=xdr

POSTGRES_HOST=localhost
POSTGRES_PORT=5432
POSTGRES_USER=admin
POSTGRES_PASSWORD=admin1234
POSTGRES_NAME=xdr
POSTGRES_SSL_MODE=disable
POSTGRES_MAX_OPEN_CONNS=10
POSTGRES_MAX_IDLE_CONNS=5
POSTGRES_CONN_MAX_LIFETIME=30m

REDIS_HOST=localhost
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_DB=0
```

#### اجرای سرور

```bash
go run cmd/web/main.go
```

سرور بر روی آدرس `http://localhost:8086` اجرا خواهد شد.

## مثال‌های استفاده

### ثبت نام کاربر

```bash
curl -X POST http://localhost:8088/api/v1/users \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "username": "nima",
    "full_name": "نیما نخستین",
    "email": "nima@example.com",
    "password": "Nima!1998"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Create successful",
  "data": null,
  "meta": null
}
```

### لاگین

```bash
curl -X POST http://localhost:8088/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "username": "nima",
    "password": "Nima!1998"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Login successful",
  "data": {
    "access_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "refresh_token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...",
    "access_token_expires_at": "2025-10-16T12:00:00Z",
    "refresh_token_expires_at": "2025-11-15T12:00:00Z"
  },
  "meta": null
}
```

### Refresh کردن Token

```bash
curl -X POST http://localhost:8088/api/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{
    "refresh_token": "YOUR_REFRESH_TOKEN"
  }'
```

**Response:**
```json
{
  "success": true,
  "message": "Tokens refreshed successful",
  "data": {
    "access_token": "NEW_ACCESS_TOKEN",
    "refresh_token": "",
    "access_token_expires_at": "2025-10-16T12:00:00Z",
    "refresh_token_expires_at": "2025-11-15T12:00:00Z"
  },
  "meta": null
}
```

### دریافت پروفایل

```bash
curl -X GET http://localhost:8088/api/v1/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "User fetched",
  "data": {
    "id": 1,
    "username": "nima",
    "full_name": "نیما نخستین",
    "email": "nima@example.com",
    "created_at": "2025-10-15T10:30:00Z",
  },
  "meta": null
}
```

### ویرایش پروفایل

```bash
curl -X PUT http://localhost:8088/api/v1/profile \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "full_name": "نیما نخستین",
    "email": "nima.new@example.com"
  }'
```

### ساخت Task جدید

```bash
curl -X POST http://localhost:8088/api/v1/tasks \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "پیاده‌سازی API احراز هویت",
    "description": "باید JWT به سیستم اضافه بشه",
    "assignee": "nima",
    "priority": "high",
    "due_date": "2025-10-20T00:00:00Z"
  }'
```

**Request Body Schema:**
```json
{
  "summary": "string (required)",
  "description": "string (optional)",
  "assignee": "string (username)",
  "priority": "lowest | low | medium | high | highest",
  "due_date": "ISO 8601 datetime"
}
```

**Response:**
```json
{
  "success": true,
  "message": "created",
  "data": null,
  "meta": null
}
```

### دریافت لیست Task‌ها

بدون فیلتر:
```bash
curl -X GET http://localhost:8088/api/v1/tasks \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

با فیلتر:
```bash
curl -X GET "http://localhost:8088/api/v1/tasks?status=ToDo&priority=high&page=1&limit=10" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Query Parameters:**
- `assignee`: نام کاربری (string)
- `status`: وضعیت (`ToDo`, `InProgress`, `Done`)
- `priority`: اولویت (`lowest`, `low`, `medium`, `high`, `highest`)
- `page`: شماره صفحه (پیش‌فرض: 1)
- `limit`: تعداد در هر صفحه (پیش‌فرض: 10)

**Response:**
```json
{
  "success": true,
  "message": "Tasks fetched successfully",
  "data": [
    {
      "id": 1,
      "summary": "پیاده‌سازی API احراز هویت",
      "description": "باید JWT به سیستم اضافه بشه",
      "assignee": {
        "id": 1,
        "username": "nima"
      },
      "status": "ToDo",
      "priority": "high",
      "due_date": "2025-10-20T00:00:00Z",
      "created_at": "2025-10-15T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total_records": 1,
    "total_pages": 1
  }
}
```

### دریافت یک Task خاص

```bash
curl -X GET http://localhost:8088/api/v1/tasks/1 \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "Task fetched successfully",
  "data": {
    "id": 1,
    "summary": "پیاده‌سازی API احراز هویت",
    "description": "باید JWT به سیستم اضافه بشه",
    "assignee": {
      "id": 1,
      "username": "nima"
    },
    "status": "ToDo",
    "priority": "high",
    "due_date": "2025-10-20T00:00:00Z",
    "created_at": "2025-10-15T10:00:00Z"
  },
  "meta": null
}
```

### ویرایش Task

```bash
curl -X PUT http://localhost:8088/api/v1/tasks/1 \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "summary": "پیاده‌سازی کامل احراز هویت",
    "description": "JWT + Refresh Token",
    "priority": "highest",
    "due_date": "2025-10-18T00:00:00Z"
  }'
```

**Request Body Schema (همه اختیاری):**
```json
{
  "summary": "string",
  "description": "string",
  "priority": "lowest | low | medium | high | highest",
  "due_date": "ISO 8601 datetime"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Task updated successfully",
  "data": null,
  "meta": null
}
```

### Assign کردن Task

```bash
curl -X PUT http://localhost:8088/api/v1/tasks/assign \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 1,
    "assignee": "ali"
  }'
```

**Request Body:**
```json
{
  "task_id": "integer (required)",
  "assignee": "string (username, required)"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Task assigned successfully",
  "data": null,
  "meta": null
}
```

### تغییر وضعیت Task

```bash
curl -X PUT http://localhost:8088/api/v1/tasks/transition \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "task_id": 1,
    "status": "InProgress"
  }'
```

**Request Body:**
```json
{
  "task_id": "integer (required)",
  "status": "ToDo | InProgress | Done (required)"
}
```

**Response:**
```json
{
  "success": true,
  "message": "Task status transitioned successfully",
  "data": null,
  "meta": null
}
```

### حذف Task

```bash
curl -X DELETE http://localhost:8088/api/v1/tasks/1 \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "Task deleted successfully",
  "data": null,
  "meta": null
}
```

### دریافت لیست کاربران (برای Assign)

```bash
curl -X GET "http://localhost:8088/api/v1/users?page=1&limit=10" \
  -H "Authorization: Bearer YOUR_ACCESS_TOKEN"
```

**Response:**
```json
{
  "success": true,
  "message": "Users fetched successfully",
  "data": [
    {
      "id": 1,
      "username": "nima",
      "full_name": "نیما احمدی",
      "email": "nima@example.com",
      "created_at": "2025-10-15T10:00:00Z",
      "updated_at": "2025-10-15T10:00:00Z"
    }
  ],
  "meta": {
    "page": 1,
    "limit": 10,
    "total_records": 1,
    "total_pages": 1
  }
}
```

## فرمت کلی Response

تمامی پاسخ‌ها از فرمت استاندارد زیر پیروی می‌کنند:

```json
{
  "success": true/false,
  "message": "پیام توضیحی",
  "data": {}, 
  "meta": {
    "page": 1,
    "limit": 10,
    "total_records": 100,
    "total_pages": 10
  }
}
```

- `success`: وضعیت موفقیت یا شکست درخواست
- `message`: پیام توضیحی
- `data`: داده‌های اصلی پاسخ (می‌تواند null باشد)
- `meta`: اطلاعات صفحه‌بندی (فقط برای لیست‌ها)

## کدهای خطا

| کد | معنی |
|----|------|
| 200 | موفقیت‌آمیز |
| 201 | ساخته شد |
| 400 | درخواست اشتباه |
| 401 | نیاز به احراز هویت |
| 404 | پیدا نشد |
| 500 | خطای سرور |

## تست‌ها

### اجرای همه تست‌ها

```bash
go test ./...
```

### اجرای تست‌ها با Coverage

```bash
go test -cover ./...
```

### اجرای تست برای یک Package خاص

```bash
go test ./domain/user/aggregate/...
```

### اجرای تست‌ها با جزئیات

```bash
go test -v ./...
```

### تولید گزارش Coverage

```bash
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out -o coverage.html
```

سپس فایل `coverage.html` را در مرورگر باز کنید.

## نکات معماری و تصمیمات طراحی

### انتخاب Redis برای Caching
Redis به عنوان لایه کش انتخاب شده است. در صورت داشتن چندین instance از API، تمامی آنها باید به یک کش مشترک دسترسی داشته باشند. Redis این نیاز را به خوبی برآورده می‌کند.

### استراتژی Invalidation کش
از روش version-based استفاده شده است. در این روش، هر بار که عملیات Create، Update یا Delete انجام می‌شود، یک شمارنده افزایش می‌یابد و تمام کش‌های قدیمی منقضی می‌شوند. این رویکرد ساده‌تر از پاک‌سازی دستی تمام کلیدها است.

### انتخاب PostgreSQL
به دلیل وجود رابطه بین Task و User، PostgreSQL انتخاب شده است. پایگاه‌داده‌های NoSQL برای این نیاز مناسب نیستند. PostgreSQL قدرتمند، سریع و معتبر است.

### استراتژی Batch Query به جای JOIN
برای نمایش اطلاعات assignee در لیست Task‌ها، به جای استفاده از JOIN، از روش Batch Query استفاده شده است. این رویکرد مزایای زیر را دارد:
- **Clean Architecture**: Separation of concerns بین domain های Task و User حفظ می‌شود
- **Cache-friendly**: امکان cache کردن مستقل اطلاعات کاربران
- **Scalable**: آماده برای مهاجرت به Microservice Architecture
- **Efficient**: در سناریوهای واقعی که assignee ها تکراری هستند، کارآمدتر است
- **Testable**: Mock کردن ساده‌تر و تست‌های مستقل‌تر

برای جزئیات بیشتر، فایل `ARCHITECTURE_DECISION.md` را مطالعه کنید.

### JWT Stateless
توکن‌ها در پایگاه داده ذخیره نمی‌شوند. این امر باعث افزایش سرعت و امکان Horizontal Scaling می‌شود. اگرچه این رویکرد معایبی دارد (مانند عدم امکان باطل کردن توکن قبل از انقضا)، اما با تعیین TTL کوتاه برای access token (24 ساعت) این مشکل کاهش می‌یابد.

### Bcrypt برای Password Hashing
از الگوریتم Bcrypt برای hash کردن رمزهای عبور استفاده شده است. این الگوریتم امن و سریع است. اگرچه Argon2 بهتر است، اما Bcrypt برای اکثر کاربردها کافی است.

### Pagination
از روش offset-based (page و limit) استفاده شده است. این روش برای مجموعه داده‌های متوسط مناسب است و کاربر می‌تواند به صفحه دلخواه خود دسترسی یابد. در صورت افزایش حجم داده‌ها، می‌توان از cursor-based pagination استفاده کرد.

## ساختار فایل‌ها

```
root/
├── cmd/
│   └── web/                # سرور اصلی
├── domain/                 # لایه Domain
│   ├── task/               # منطق Task
│   └── user/               # منطق User
├── interfaces/             # لایه Presentation
│   └── http/
│       ├── handlers/       # هندلرهای HTTP
│       ├── middleware/     # میدلور (احراز هویت و...)
│       └── server/         # راه‌اندازی سرور
├── services/               # لایه Application
│   ├── task/               # سرویس Task
│   └── user/               # سرویس User
├── pkg/                    # Infrastructure
│   ├── jwt/                # مدیریت Token
│   ├── postgres/           # کلاینت دیتابیس
│   ├── redis/              # کلاینت Redis
│   └── response/           # پاسخ‌های استاندارد
└── docs/                   # مستندات Swagger
```

### مسئولیت لایه‌ها

**Domain Layer (`domain/`):**
- منطق کسب‌وکار اصلی
- Entity ها و Value Object ها
- Aggregate ها (مانند TaskResponse برای ترکیب اطلاعات از چند entity)
- Interface های Repository
- بدون وابستگی به سایر لایه‌ها

**Service Layer (`services/`):**
- پیاده‌سازی Use Case ها
- مدیریت منطق‌های پیچیده
- فراخوانی Repository ها
- مدیریت کش

**Interface Layer (`interfaces/`):**
- هندلرهای HTTP
- تبدیل Request/Response
- اعتبارسنجی ورودی
- میدلورها

**Infrastructure Layer (`pkg/`):**
- کلاینت‌های پایگاه داده
- کلاینت Redis
- مدیریت JWT
- مدیریت تنظیمات

## امنیت

ملاحظات امنیتی رعایت شده:

1. **JWT Secret**: استفاده از کلیدهای قوی (حداقل 32 کاراکتر)
2. **Password**: Hash شدن با Bcrypt (cost factor: 10)
3. **SQL Injection**: استفاده از GORM با parameterized query
4. **HTTPS**: الزامی در محیط production
5. **Environment Variables**: عدم commit فایل `.env` در git
6. **Token Expiry**: access token پس از 24 ساعت، refresh token پس از 30 روز منقضی می‌شود

## مانیتورینگ

در صورت اجرای docker-compose، سرویس‌های Prometheus و Grafana نیز فعال می‌شوند:

- **Prometheus**: `http://localhost:9090`
- **Grafana**: `http://localhost:3000` (admin/admin)

متریک‌های سیستم مانند تعداد درخواست‌ها، زمان پاسخ‌دهی و سایر معیارهای عملکردی قابل مشاهده هستند.

---

**توسعه داده شده با Go، Gin، PostgreSQL و Redis**

