#!/bin/bash
# User data script for URL Shortener application

# Update system
yum update -y

# Install Docker
yum install -y docker
systemctl start docker
systemctl enable docker
usermod -a -G docker ec2-user

# Install Docker Compose
curl -L "https://github.com/docker/compose/releases/latest/download/docker-compose-$(uname -s)-$(uname -m)" -o /usr/local/bin/docker-compose
chmod +x /usr/local/bin/docker-compose

# Install Git
yum install -y git

# Clone the repository (you'll need to update this with your actual repository URL)
cd /home/ec2-user
# git clone https://github.com/yourusername/url-shortener.git
# For now, we'll create a simple deployment structure

# Create application directory
mkdir -p /opt/url-shortener
cd /opt/url-shortener

# Create environment file
cat > .env << EOF
# Database configuration
DB_HOST=${db_host}
DB_PORT=5432
DB_NAME=${db_name}
DB_USER=${db_user}
DB_PASSWORD=${db_password}

# Redis configuration
REDIS_ADDR=${redis_endpoint}:6379
REDIS_PASSWORD=${redis_auth}

# Application configuration
PORT=${app_port}
ENVIRONMENT=production

# Other configurations
BASE_URL=http://$(curl -s http://169.254.169.254/latest/meta-data/public-hostname)
EOF

# Create a simple systemd service file
cat > /etc/systemd/system/url-shortener.service << EOF
[Unit]
Description=URL Shortener Application
After=docker.service
Requires=docker.service

[Service]
Type=oneshot
RemainAfterExit=yes
WorkingDirectory=/opt/url-shortener
ExecStart=/usr/local/bin/docker-compose up -d
ExecStop=/usr/local/bin/docker-compose down
User=root

[Install]
WantedBy=multi-user.target
EOF

# Create docker-compose.yml for production
cat > docker-compose.yml << EOF
version: '3.8'

services:
  app:
    # You'll need to build and push your image to a registry (Docker Hub, ECR, etc.)
    # image: your-registry/url-shortener:latest
    build: .
    ports:
      - "${app_port}:${app_port}"
    environment:
      - DB_HOST=\${DB_HOST}
      - DB_PORT=\${DB_PORT}
      - DB_NAME=\${DB_NAME}
      - DB_USER=\${DB_USER}
      - DB_PASSWORD=\${DB_PASSWORD}
      - REDIS_ADDR=\${REDIS_ADDR}
      - REDIS_PASSWORD=\${REDIS_PASSWORD}
      - PORT=\${PORT}
      - ENVIRONMENT=\${ENVIRONMENT}
      - BASE_URL=\${BASE_URL}
    restart: unless-stopped
    depends_on:
      - migrate

  migrate:
    # Same image as app, but run migrations
    # image: your-registry/url-shortener:latest
    build: .
    command: ["/app/migrate"]
    environment:
      - DB_HOST=\${DB_HOST}
      - DB_PORT=\${DB_PORT}
      - DB_NAME=\${DB_NAME}
      - DB_USER=\${DB_USER}
      - DB_PASSWORD=\${DB_PASSWORD}
    restart: "no"
EOF

# Set proper permissions
chown -R ec2-user:ec2-user /opt/url-shortener
chmod +x /opt/url-shortener

# Install CloudWatch agent (optional, for monitoring)
yum install -y amazon-cloudwatch-agent

# Create log directory
mkdir -p /var/log/url-shortener
chown ec2-user:ec2-user /var/log/url-shortener

# Enable and start the service
systemctl daemon-reload
systemctl enable url-shortener.service

# Note: You'll need to manually deploy your application code
# This can be done through:
# 1. GitHub Actions deployment
# 2. AWS CodeDeploy
# 3. Manual deployment via SSH
echo "User data script completed. Please deploy your application code manually or through CI/CD pipeline."
