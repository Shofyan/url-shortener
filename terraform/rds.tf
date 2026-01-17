# DB Subnet Group
resource "aws_db_subnet_group" "main" {
  name       = "${var.project_name}-db-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name        = "${var.project_name}-db-subnet-group"
    Environment = var.environment
  }
}

# RDS PostgreSQL Database
resource "aws_db_instance" "postgres" {
  identifier = "${var.project_name}-postgres"

  # Engine configuration
  engine         = "postgres"
  engine_version = "15.7"
  instance_class = "db.t3.micro" # Free tier eligible

  # Storage configuration
  allocated_storage     = 20 # Free tier allows up to 20 GB
  max_allocated_storage = 20
  storage_type          = "gp2"
  storage_encrypted     = false # Free tier doesn't support encryption

  # Database configuration
  db_name  = "urlshortener"
  username = var.db_username
  password = var.db_password

  # Network configuration
  db_subnet_group_name   = aws_db_subnet_group.main.name
  vpc_security_group_ids = [aws_security_group.rds.id]
  publicly_accessible    = false

  # Backup configuration
  backup_retention_period = 7
  backup_window          = "03:00-04:00"
  maintenance_window     = "sun:04:00-sun:05:00"

  # Other configurations
  skip_final_snapshot = true
  deletion_protection = false

  tags = {
    Name        = "${var.project_name}-postgres"
    Environment = var.environment
  }
}
