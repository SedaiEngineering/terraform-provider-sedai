# Minimal — group containing all AWS EC2 instances in us-east-1
resource "sedai_create_group" "ec2_us_east_1" {
  name          = "ec2-us-east-1"
  resource_type = ["AWS_EC2"]
  region        = ["us-east-1"]
}

# Tag-filtered — values containing '*' are sent as regex matchers,
# everything else is sent as an exact matcher
resource "sedai_create_group" "db_backend_apps" {
  name         = "db-backend-apps"
  auto_refresh = true

  tags = {
    app = ["db-backend", "api-*"]
    env = ["prod"]
  }

  resource_type = ["AWS_EC2", "AWS_LAMBDA"]
}

# Cluster-scoped — pin the group to a single ECS cluster
resource "sedai_create_group" "labs_ecs_cluster" {
  name = "labs-ecs-cluster"

  cluster = [
    "arn:aws:ecs:us-east-1:000000000000:cluster/sedai-labs-ecs-02",
  ]

  resource_type = ["AWS_ECS"]
}

# Kubernetes — filter by namespace and workload type
resource "sedai_create_group" "k8s_payments_deployments" {
  name = "payments-deployments"

  namespace     = ["payments", "payments-staging"]
  resource_type = ["KUBERNETES_DEPLOYMENT", "KUBERNETES_STATEFULSET"]
}

# Subgroup — parent_group_id links this group under an existing parent
resource "sedai_create_group" "payments_us_east_1" {
  name            = "payments-us-east-1"
  parent_group_id = sedai_create_group.k8s_payments_deployments.id
  region          = ["us-east-1"]
}
