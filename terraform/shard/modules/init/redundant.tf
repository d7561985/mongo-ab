
# Create a VPC
#resource "aws_vpc" "mongo" {
#  cidr_block = "10.0.0.0/16"
#  tags       = {
#    Name = "mongodb-${var.ENVIRONMENT}"
#  }
#}

#resource "aws_subnet" "mongo_main" {
#  vpc_id     = aws_vpc.mongo.id
#  cidr_block = "10.0.1.0/24"
#  availability_zone = "${var.AWS_REGION}a"
#
#  # make it public
#  map_public_ip_on_launch = true
#
#  tags = {
#    Name = "mongodb-${var.ENVIRONMENT}"
#    Type: "public"
#  }
#}

#resource "aws_spot_fleet_request" "example" {
#  count = 0 # disable
#  iam_fleet_role  = var.FLEET_ROLE
#  spot_price      = var.SPOT_PRICE
#  target_capacity = 2
#
#  launch_template_config {
#    launch_template_specification {
#      id      = aws_launch_template.mongo.id
#      version = aws_launch_template.mongo.latest_version
#    }
#
#    overrides {
#      availability_zone = "${var.AWS_REGION}a"
#      subnet_id = data.aws_subnet.default.id
#      spot_price = var.SPOT_PRICE
#    }
#  }
#
#  depends_on = [aws_launch_template.mongo]
#}
