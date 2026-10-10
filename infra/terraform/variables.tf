variable "aws_region" {
  description = "AWS region"
  type        = string
  default     = "ap-south-1"
}

variable "project" {
  description = "Project name prefix for all resources"
  type        = string
  default     = "rate-limiter"
}

variable "core_image" {
  description = "Docker image for the Go core service"
  type        = string
  default     = "ghcr.io/Anand09-in/RateLimiter/core:latest"
}

variable "db_name" {
  description = "PostgreSQL database name"
  type        = string
  default     = "ratelimiter"
}

variable "db_user" {
  description = "PostgreSQL username"
  type        = string
  default     = "rluser"
}

variable "db_password" {
  description = "PostgreSQL password"
  type        = string
  sensitive   = true
}
