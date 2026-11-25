# Account-wide default for Kubernetes workloads. is_prod may stay in the
# block only if it agrees with the mode (true with PROD, false otherwise).
resource "sedai_account_settings" "eks_prod" {
  account_id        = sedai_account.eks_prod.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  kube_app_settings {
    aggression_mode = "PROD"
    is_prod         = true
  }
}
