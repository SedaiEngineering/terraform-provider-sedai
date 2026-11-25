# Turbo for the group's Kubernetes workloads, Expedite for its ECS services.
# A mode change is applied as a switch first (it resets the CPU and memory
# floors to the mode's defaults); the floors set here are saved after it.
resource "sedai_group_settings" "payments_nonprod" {
  group_id          = sedai_group.payments_nonprod.id
  availability_mode = "DATA_PILOT"
  optimization_mode = "DATA_PILOT"

  kube_app_settings {
    aggression_mode                = "TURBO" # PROD, EXPEDITE or TURBO
    vertical_scaling_min_cpu_cores = 0.25
  }

  ecs_app_settings {
    aggression_mode = "EXPEDITE" # PROD or EXPEDITE; TURBO is Kubernetes only
  }
}
