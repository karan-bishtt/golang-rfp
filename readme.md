rfp-management-system/
├── auth-service/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── controllers/
│   │   │   └── auth_controller.go
│   │   ├── models/
│   │   │   └── user.go
│   │   ├── routes/
│   │   │   └── auth_routes.go
│   │   ├── middleware/
│   │   │   └── auth_middleware.go
│   │   ├── utils/
│   │   │   ├── jwt.go
│   │   │   └── password.go
│   │   └── database/
│   │       └── connection.go
│   ├── config/
│   │   └── config.go
│   ├── go.mod
│   └── go.sum
│
├── notification-service/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── controllers/
│   │   │   └── notification_controller.go
│   │   ├── models/
│   │   │   └── notification.go
│   │   ├── routes/
│   │   │   └── notification_routes.go
│   │   ├── middleware/
│   │   │   └── auth_middleware.go
│   │   ├── utils/
│   │   │   ├── email.go
│   │   │   └── sms.go
│   │   └── database/
│   │       └── connection.go
│   ├── config/
│   │   └── config.go
│   ├── go.mod
│   └── go.sum
│
├── category-service/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── controllers/
│   │   │   └── category_controller.go
│   │   ├── models/
│   │   │   └── category.go
│   │   ├── routes/
│   │   │   └── category_routes.go
│   │   ├── middleware/
│   │   │   └── auth_middleware.go
│   │   ├── utils/
│   │   │   └── validator.go
│   │   └── database/
│   │       └── connection.go
│   ├── config/
│   │   └── config.go
│   ├── go.mod
│   └── go.sum
│
├── rfp-quote-service/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── controllers/
│   │   │   ├── rfp_controller.go
│   │   │   └── quote_controller.go
│   │   ├── models/
│   │   │   ├── rfp.go
│   │   │   └── quote.go
│   │   ├── routes/
│   │   │   ├── rfp_routes.go
│   │   │   └── quote_routes.go
│   │   ├── middleware/
│   │   │   └── auth_middleware.go
│   │   ├── utils/
│   │   │   ├── file_upload.go
│   │   │   └── validator.go
│   │   └── database/
│   │       └── connection.go
│   ├── config/
│   │   └── config.go
│   ├── go.mod
│   └── go.sum
│
├── user-service/
│   ├── cmd/
│   │   └── main.go
│   ├── internal/
│   │   ├── controllers/
│   │   │   └── user_controller.go
│   │   ├── models/
│   │   │   └── user.go
│   │   ├── routes/
│   │   │   └── user_routes.go
│   │   ├── middleware/
│   │   │   └── auth_middleware.go
│   │   ├── utils/
│   │   │   ├── validator.go
│   │   │   └── profile.go
│   │   └── database/
│   │       └── connection.go
│   ├── config/
│   │   └── config.go
│   ├── go.mod
│   └── go.sum
│
├── docker-compose.yml
└── README.md




go.mod File
Purpose:

Module definition - Defines your project as a Go module
Dependency management - Lists direct dependencies and their versions
Go version requirement - Specifies minimum Go version needed


go.sum FilePurpose:

Security verification - Contains cryptographic checksums
Reproducible builds - Ensures exact same versions are used
Integrity checking - Prevents tampering with dependencies


<!-- # Production mode
make up

# Development mode with debugging
make debug

# All services
make logs

# Specific service
make logs-auth-service
make logs-category-service

# Access database shell
make db-shell

# Connect to specific database
docker-compose exec postgres psql -U postgres -d auth_db
 -->


1️⃣ What docker compose up -d does

When you run:

docker compose up -d


Docker Compose does both:

Builds images (if a build: context is specified in your docker-compose.yml) or pulls them (if image: is specified).

Runs containers from those images.

If the image already exists and hasn't changed, Compose won't rebuild it unless you explicitly tell it.

If containers are already running, up will try to reuse them, unless changes require a rebuild.


3️⃣ Commands to apply the env variables
Option 1: Recreate containers without rebuilding images
docker compose up -d


If containers are already running, Compose may not replace them.

To force recreation:

docker compose up -d --force-recreate


This will:

Stop existing containers

Recreate them with the new environment variables

Keep using existing images (no rebuild)

✅ Use this if you only changed env variables and not Dockerfile or source code.


Option 2: Rebuild only if necessary (like Go code changes)
docker compose up -d --build


This rebuilds images (Go services) and recreates containers.

Not needed if you only changed env variables.

Option 3: Just recreate container (manual)
docker compose stop <service_name>
docker compose rm -f <service_name>
docker compose up -d <service_name>


This stops, removes, and recreates the container with updated env variables.

Still uses existing image, no rebuild needed.


✅ TL;DR

You do not need to rebuild if you only updated env vars.

Command you want:

docker compose up -d --force-recreate


This will apply new environment variables to your containers without touching your Postgres image or rebuilding Go services unnecessarily.


NGNIX


Use Nginx as reverse proxy (Recommended)

Install Nginx:

sudo apt install -y nginx
sudo systemctl enable nginx
sudo systemctl start nginx


Configure reverse proxy for your services:

sudo nano /etc/nginx/sites-available/golang-rfp


Example config:

server {
    listen 80;
    server_name golang-rfp.veldev.com;

    location /auth/ {
        proxy_pass http://localhost:8081/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /category/ {
        proxy_pass http://localhost:8083/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /notification/ {
        proxy_pass http://localhost:8082/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }

    location /rfp-quote/ {
        proxy_pass http://localhost:8084/;
        proxy_set_header Host $host;
        proxy_set_header X-Real-IP $remote_addr;
    }
}


Enable the site:

sudo ln -s /etc/nginx/sites-available/golang-rfp /etc/nginx/sites-enabled/
sudo nginx -t
sudo systemctl reload nginx


Now your services are accessible via paths like:

http://golang-rfp.veldev.com/auth/
http://golang-rfp.veldev.com/category/
http://golang-rfp.veldev.com/quote/

