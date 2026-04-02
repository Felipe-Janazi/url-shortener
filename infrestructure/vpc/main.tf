# ── VPC ───────────────────────────────────────────────────────────────────────
# Cria a rede privada virtual onde todos os recursos da aplicação vão residir.
# A separação em subnets públicas/privadas é essencial para segurança:
# - Pública: ALB (recebe tráfego da internet)
# - Privada: ECS, RDS, ElastiCache (nunca expostos diretamente)

terraform {
  required_providers {
    aws = { source = "hashicorp/aws", version = "~> 5.0" }
  }
}

variable "project"    { default = "url-shortener" }
variable "aws_region" { default = "us-east-1" }

# Busca as Availability Zones disponíveis na região para distribuir os recursos.
# Usar pelo menos 2 AZs garante alta disponibilidade — se uma cair, a outra continua.
data "aws_availability_zones" "available" {}

resource "aws_vpc" "main" {
  cidr_block           = "10.0.0.0/16"   # 65.536 endereços disponíveis
  enable_dns_support   = true             # necessário para o RDS resolver nomes internos
  enable_dns_hostnames = true

  tags = { Name = "${var.project}-vpc" }
}

# ── Subnets públicas (ALB) ────────────────────────────────────────────────────
# Cada subnet pública fica em uma AZ diferente para redundância.
resource "aws_subnet" "public" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  # Instâncias nessa subnet recebem IP público automaticamente.
  # Necessário para o ALB ser acessível da internet.
  map_public_ip_on_launch = true

  tags = { Name = "${var.project}-public-${count.index}" }
}

# ── Subnets privadas (ECS, RDS, Redis) ───────────────────────────────────────
resource "aws_subnet" "private" {
  count             = 2
  vpc_id            = aws_vpc.main.id
  cidr_block        = "10.0.${count.index + 10}.0/24"
  availability_zone = data.aws_availability_zones.available.names[count.index]

  tags = { Name = "${var.project}-private-${count.index}" }
}

# Internet Gateway: porta de entrada/saída para a internet nas subnets públicas.
resource "aws_internet_gateway" "main" {
  vpc_id = aws_vpc.main.id
  tags   = { Name = "${var.project}-igw" }
}

# Route table pública: todo tráfego não-local vai para o Internet Gateway.
resource "aws_route_table" "public" {
  vpc_id = aws_vpc.main.id
  route {
    cidr_block = "0.0.0.0/0"
    gateway_id = aws_internet_gateway.main.id
  }
  tags = { Name = "${var.project}-public-rt" }
}

resource "aws_route_table_association" "public" {
  count          = 2
  subnet_id      = aws_subnet.public[count.index].id
  route_table_id = aws_route_table.public.id
}

# ── Outputs (usados pelos outros módulos) ─────────────────────────────────────
output "vpc_id"             { value = aws_vpc.main.id }
output "public_subnet_ids"  { value = aws_subnet.public[*].id }
output "private_subnet_ids" { value = aws_subnet.private[*].id }