aws_region             = "eu-west-1"
env                    = "dev"
vpc_cidr               = "10.0.0.0/16"
public_subnets         = [ "10.0.0.0/20", "10.0.32.0/20" ]
api_name               = "payment-api"
container_port         = 8000
repository             = "payments-api"
image_tag              = "v1"
db_storage             = 20
db_pass                = "test123456"
db_engine              = "postgres"
db_instance_class      = "db.t2.micro"
db_name                = "payments"
db_username            = "api"
db_password            = "test123456"
db_engine_version      = "10.6"
db_port                = "5432"