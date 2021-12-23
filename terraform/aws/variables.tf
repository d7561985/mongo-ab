variable "shardDB" {
  type = string
  description = "database which will use sharding"
  default = "db"
}

variable "useNvME" {
  type = bool
  default = true
}