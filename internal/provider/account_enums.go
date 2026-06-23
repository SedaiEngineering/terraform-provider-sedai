package provider

// Provider-side enum lists for sedai_account schema validation.
// These live here (not in the SDK) because they are Terraform UX constraints,
// not SDK domain logic — the SDK does not validate these fields.

var validCloudProviders = []string{"AWS", "AZURE", "GCP", "KUBERNETES"}

var validIntegrationTypes = []string{"AGENTLESS", "AGENT_BASED"}

var validClusterProviders = []string{"AWS", "GCP", "AZURE", "SELF_MANAGED"}
