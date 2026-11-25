# One Kubernetes workload on Turbo. For an ECS resource use PROD or EXPEDITE.
resource "sedai_resource_settings" "checkout_api" {
  resource_id       = "abc123/Deployment/prod-cluster/payments/checkout-api"
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"
  aggression_mode   = "TURBO"
}
