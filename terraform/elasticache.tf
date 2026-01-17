# ElastiCache Subnet Group
resource "aws_elasticache_subnet_group" "main" {
  name       = "${var.project_name}-cache-subnet-group"
  subnet_ids = aws_subnet.private[*].id

  tags = {
    Name        = "${var.project_name}-cache-subnet-group"
    Environment = var.environment
  }
}

# ElastiCache Redis Cluster
resource "aws_elasticache_replication_group" "redis" {
  replication_group_id       = "${var.project_name}-redis"
  description                = "Redis cluster for ${var.project_name}"

  # Node configuration
  node_type               = "cache.t2.micro" # Free tier eligible
  port                    = 6379
  parameter_group_name    = "default.redis7"

  # Cluster configuration
  num_cache_clusters      = 1

  # Security and network configuration
  subnet_group_name       = aws_elasticache_subnet_group.main.name
  security_group_ids      = [aws_security_group.redis.id]

  # Authentication
  auth_token              = var.redis_auth_token != "" ? var.redis_auth_token : null
  transit_encryption_enabled = var.redis_auth_token != "" ? true : false
  at_rest_encryption_enabled = false # Free tier doesn't support encryption at rest

  # Backup configuration
  snapshot_retention_limit = 0 # Free tier doesn't support automated backups

  tags = {
    Name        = "${var.project_name}-redis"
    Environment = var.environment
  }
}
