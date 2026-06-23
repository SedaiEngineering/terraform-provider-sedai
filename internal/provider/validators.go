package provider

import (
	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/constants"
	"github.com/hashicorp/terraform-plugin-framework-validators/int64validator"
	"github.com/hashicorp/terraform-plugin-framework-validators/listvalidator"
	"github.com/hashicorp/terraform-plugin-framework-validators/stringvalidator"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// This file is the single source of truth for plan-time validators reused
// across every Sedai resource and data source. Each validator is sourced
// from `constants` so updating an enum there propagates to every schema
// automatically — no place for the validator and the constants to drift
// apart.

// settingsConfigModeValidator returns a String-attribute validator that
// rejects values not in constants.ValidSettingsConfigModes. Used by every
// `availability_mode` and `optimization_mode` attribute across group /
// account / resource settings resources.
func settingsConfigModeValidator() validator.String {
	return stringvalidator.OneOf(validSettingsConfigModeStrings()...)
}

// validSettingsConfigModeStrings flattens the typed constants slice into
// plain strings so it can be passed to stringvalidator.OneOf.
func validSettingsConfigModeStrings() []string {
	out := make([]string, 0, len(constants.ValidSettingsConfigModes))
	for _, m := range constants.ValidSettingsConfigModes {
		out = append(out, string(m))
	}
	return out
}

// groupResourceTypeListValidator returns a List-attribute validator that
// rejects any list element not in the accepted resource-type set. Used by
// `sedai_group.resource_types`. Includes the user-facing
// `KUBERNETES_DAEMONSET` spelling alongside the constants list (the typed
// constants store the backend's misspelled `KUBERNETES_DEAMONSET` form;
// HCL accepts either and `buildGroupDefinition` translates).
func groupResourceTypeListValidator() validator.List {
	return listvalidator.ValueStringsAre(
		stringvalidator.OneOf(validGroupResourceTypeStrings()...),
	)
}

// validGroupResourceTypeStrings is the accepted-values list for HCL —
// constants.ValidGroupResourceTypes plus the correctly-spelled
// `KUBERNETES_DAEMONSET` alias.
func validGroupResourceTypeStrings() []string {
	out := make([]string, 0, len(constants.ValidGroupResourceTypes)+1)
	for _, t := range constants.ValidGroupResourceTypes {
		out = append(out, string(t))
	}
	out = append(out, "KUBERNETES_DAEMONSET")
	return out
}

// priorityValidator returns an Int64-attribute validator that rejects
// priority values below 1. Priority is 1-based across the SDK and TF
// surface (the spec mandates this; the SDK translates to 0-based for the
// wire). Use on every `sedai_group_priority.group_priorities.priority`
// field.
func priorityValidator() validator.Int64 {
	return int64validator.AtLeast(1)
}

// noAutoConfigModeValidator returns a String-attribute validator that
// accepts `DATA_PILOT` and `CO_PILOT` only — the spec's restriction for:
//   - bucket_settings.optimization_mode (S3 doesn't support AUTO)
//   - volume_settings.availability_mode
//   - volume_settings.optimization_mode  (EBS doesn't support AUTO)
//   - app_settings.availability_mode / .optimization_mode
//
// Excludes the third value (`AUTO`) without hardcoding the strings:
// derived from constants.ValidSettingsConfigModes so adding a new mode
// there automatically propagates.
func noAutoConfigModeValidator() validator.String {
	all := validSettingsConfigModeStrings()
	noAuto := make([]string, 0, len(all))
	for _, m := range all {
		if m != "AUTO" {
			noAuto = append(noAuto, m)
		}
	}
	return stringvalidator.OneOf(noAuto...)
}

// noCopilotConfigModeValidator returns a String-attribute validator that
// accepts `DATA_PILOT` and `AUTO` only — the spec's restriction for:
//   - serverless_settings.optimization_mode (Lambda doesn't expose CO_PILOT)
func noCopilotConfigModeValidator() validator.String {
	all := validSettingsConfigModeStrings()
	noCopilot := make([]string, 0, len(all))
	for _, m := range all {
		if m != "CO_PILOT" {
			noCopilot = append(noCopilot, m)
		}
	}
	return stringvalidator.OneOf(noCopilot...)
}

// optimizationFocusValidator returns a String-attribute validator for
// optimization_focus fields. Accepted values come from
// constants.ValidOptimizationFocuses so the schema and constants stay
// in lockstep.
func optimizationFocusValidator() validator.String {
	out := make([]string, 0, len(constants.ValidOptimizationFocuses))
	for _, f := range constants.ValidOptimizationFocuses {
		out = append(out, string(f))
	}
	return stringvalidator.OneOf(out...)
}

// concurrencyModeValidator returns a String-attribute validator for
// serverless_settings.concurrency_mode. Accepted values come from
// constants.ValidConcurrencyModes.
func concurrencyModeValidator() validator.String {
	out := make([]string, 0, len(constants.ValidConcurrencyModes))
	for _, m := range constants.ValidConcurrencyModes {
		out = append(out, string(m))
	}
	return stringvalidator.OneOf(out...)
}

// cloudProviderValidator rejects cloud_provider values not in validCloudProviders.
func cloudProviderValidator() validator.String {
	return stringvalidator.OneOf(validCloudProviders...)
}

// integrationTypeValidator rejects integration_type values not in validIntegrationTypes.
func integrationTypeValidator() validator.String {
	return stringvalidator.OneOf(validIntegrationTypes...)
}

// clusterProviderValidator rejects cluster_provider values not in validClusterProviders.
func clusterProviderValidator() validator.String {
	return stringvalidator.OneOf(validClusterProviders...)
}

// managedServicesValidator rejects any list element not in the union of all
// cloud managed-service values. Cross-cloud validation (e.g. AWS-only values
// rejected for an Azure account) is deferred to ticket 11 cross-field validators.
func managedServicesValidator() validator.List {
	all := make([]string, 0)
	for _, services := range constants.SupportedManagedServices {
		for _, s := range services {
			all = append(all, string(s))
		}
	}
	return listvalidator.ValueStringsAre(stringvalidator.OneOf(all...))
}
