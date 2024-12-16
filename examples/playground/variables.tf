variable "mongodb_user" {
  description = "MongoDB username"
  sensitive   = true
  default = "test"
}

variable "mongodb_password" {
  description = "MongoDB password"
  sensitive   = true
   default = "test"
}

variable "mongodb_host" {
  description = "MongoDB host"
   default = "test"
}

variable "postgres_user" {
  description = "PostgreSQL username"
  sensitive   = true
   default = "test"
}

variable "postgres_password" {
  description = "PostgreSQL password"
  sensitive   = true
   default = "test"
}

variable "postgres_host" {
  description = "PostgreSQL host"
   default = "test"
}

variable "postgres_db" {
  description = "PostgreSQL database name"
   default = "test"
}
