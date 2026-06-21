package provider

// This file contains data source schema block builders for the 7 per-resource-type
// settings blocks. These are separate from the resource/schema versions in each
// *_settings.go file because datasource/schema and resource/schema are distinct
// Go packages with incompatible types — all attributes here are Computed: true.

import (
	dsschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
)

func dataSourceKubeAppSettingsBlock() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		Description: "Per-resource-type settings for Kubernetes workloads.",
		Attributes: map[string]dsschema.Attribute{
			"availability_mode":                     dsschema.StringAttribute{Computed: true},
			"optimization_mode":                     dsschema.StringAttribute{Computed: true},
			"optimization_focus":                    dsschema.StringAttribute{Computed: true},
			"sedai_sync_enabled":                    dsschema.BoolAttribute{Computed: true},
			"is_prod":                               dsschema.BoolAttribute{Computed: true},
			"is_operation_allowed":                  dsschema.BoolAttribute{Computed: true},
			"horizontal_scaling_enabled":            dsschema.BoolAttribute{Computed: true},
			"horizontal_scaling_min_replicas":       dsschema.Int64Attribute{Computed: true},
			"horizontal_scaling_max_replicas":       dsschema.Int64Attribute{Computed: true},
			"horizontal_scaling_replica_multiplier": dsschema.Int64Attribute{Computed: true},
			"vertical_scaling_enabled":              dsschema.BoolAttribute{Computed: true},
			"vertical_scaling_min_cpu_cores":        dsschema.Float64Attribute{Computed: true},
			"vertical_scaling_min_memory_bytes":     dsschema.Int64Attribute{Computed: true},
			"predictive_scaling_enabled":            dsschema.BoolAttribute{Computed: true},
			"autonomous_action_without_traffic":     dsschema.BoolAttribute{Computed: true},
			"max_latency_increase_pct":              dsschema.Int64Attribute{Computed: true},
			"max_cpu_increase_pct":                  dsschema.Int64Attribute{Computed: true},
			"max_memory_increase_pct":               dsschema.Int64Attribute{Computed: true},
		},
	}
}

func dataSourceBucketSettingsBlock() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		Description: "Per-resource-type settings for S3 buckets.",
		Attributes: map[string]dsschema.Attribute{
			"optimization_mode": dsschema.StringAttribute{Computed: true},
			"sedai_sync_enabled": dsschema.BoolAttribute{Computed: true},
		},
	}
}

func dataSourceAppSettingsBlock() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		Description: "Per-resource-type settings for generic application workloads.",
		Attributes: map[string]dsschema.Attribute{
			"availability_mode":               dsschema.StringAttribute{Computed: true},
			"optimization_mode":               dsschema.StringAttribute{Computed: true},
			"is_prod":                         dsschema.BoolAttribute{Computed: true},
			"sedai_sync_enabled":              dsschema.BoolAttribute{Computed: true},
			"horizontal_scaling_enabled":      dsschema.BoolAttribute{Computed: true},
			"horizontal_scaling_min_replicas": dsschema.Int64Attribute{Computed: true},
			"horizontal_scaling_max_replicas": dsschema.Int64Attribute{Computed: true},
		},
	}
}

func dataSourceContainerAppSettingsBlock() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		Description: "Per-resource-type settings for containerized workloads.",
		Attributes: map[string]dsschema.Attribute{
			"availability_mode":               dsschema.StringAttribute{Computed: true},
			"optimization_mode":               dsschema.StringAttribute{Computed: true},
			"is_prod":                         dsschema.BoolAttribute{Computed: true},
			"sedai_sync_enabled":              dsschema.BoolAttribute{Computed: true},
			"horizontal_scaling_enabled":      dsschema.BoolAttribute{Computed: true},
			"horizontal_scaling_min_replicas": dsschema.Int64Attribute{Computed: true},
			"horizontal_scaling_max_replicas": dsschema.Int64Attribute{Computed: true},
			"vertical_scaling_enabled":        dsschema.BoolAttribute{Computed: true},
			"predictive_scaling_enabled":      dsschema.BoolAttribute{Computed: true},
			"optimization_focus":              dsschema.StringAttribute{Computed: true},
			"max_latency_increase_pct":        dsschema.Int64Attribute{Computed: true},
			"max_cpu_increase_pct":            dsschema.Int64Attribute{Computed: true},
			"max_memory_increase_pct":         dsschema.Int64Attribute{Computed: true},
			"autonomous_action_without_traffic": dsschema.BoolAttribute{Computed: true},
		},
	}
}

func dataSourceECSAppSettingsBlock() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		Description: "Per-resource-type settings for AWS ECS service workloads.",
		Attributes: map[string]dsschema.Attribute{
			"availability_mode":                    dsschema.StringAttribute{Computed: true},
			"optimization_mode":                    dsschema.StringAttribute{Computed: true},
			"is_prod":                              dsschema.BoolAttribute{Computed: true},
			"sedai_sync_enabled":                   dsschema.BoolAttribute{Computed: true},
			"horizontal_scaling_enabled":           dsschema.BoolAttribute{Computed: true},
			"horizontal_scaling_min_replicas":      dsschema.Int64Attribute{Computed: true},
			"horizontal_scaling_max_replicas":      dsschema.Int64Attribute{Computed: true},
			"vertical_scaling_enabled":             dsschema.BoolAttribute{Computed: true},
			"predictive_scaling_enabled":           dsschema.BoolAttribute{Computed: true},
			"optimization_focus":                   dsschema.StringAttribute{Computed: true},
			"max_latency_increase_pct":             dsschema.Int64Attribute{Computed: true},
			"max_cpu_increase_pct":                 dsschema.Int64Attribute{Computed: true},
			"max_memory_increase_pct":              dsschema.Int64Attribute{Computed: true},
			"autonomous_action_without_traffic":    dsschema.BoolAttribute{Computed: true},
			"service_autoscaling_enabled":          dsschema.BoolAttribute{Computed: true},
			"horizontal_scaling_replica_increment": dsschema.Int64Attribute{Computed: true},
			"vertical_scaling_min_cpu":             dsschema.Int64Attribute{Computed: true},
			"vertical_scaling_min_memory":          dsschema.Int64Attribute{Computed: true},
		},
	}
}

func dataSourceServerlessSettingsBlock() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		Description: "Per-resource-type settings for AWS Lambda functions.",
		Attributes: map[string]dsschema.Attribute{
			"availability_mode":    dsschema.StringAttribute{Computed: true},
			"optimization_mode":    dsschema.StringAttribute{Computed: true},
			"optimization_focus":   dsschema.StringAttribute{Computed: true},
			"concurrency_mode":     dsschema.StringAttribute{Computed: true},
			"max_cost_change_pct":  dsschema.Int64Attribute{Computed: true},
			"max_latency_change_pct": dsschema.Int64Attribute{Computed: true},
			"sedai_sync_enabled":   dsschema.BoolAttribute{Computed: true},
		},
	}
}

func dataSourceVolumeSettingsBlock() dsschema.SingleNestedBlock {
	return dsschema.SingleNestedBlock{
		Description: "Per-resource-type settings for AWS EBS volumes.",
		Attributes: map[string]dsschema.Attribute{
			"availability_mode": dsschema.StringAttribute{Computed: true},
			"optimization_mode": dsschema.StringAttribute{Computed: true},
			"sedai_sync_enabled": dsschema.BoolAttribute{Computed: true},
		},
	}
}
