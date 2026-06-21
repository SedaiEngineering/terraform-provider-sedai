package provider

import (
	"time"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/monitoringProvider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// verifyMonitoringProviderCreated polls GET after a failed POST to check if the backend
// created the monitoring provider despite the connection error (EOF-during-POST).
// Returns the existing provider's map if found, nil otherwise.
// See LIMITATIONS.md for known edge cases with pre-existing resources.
func verifyMonitoringProviderCreated(accountId, mpType string) map[string]interface{} {
	for i := 0; i < 3; i++ {
		time.Sleep(2 * time.Second)
		providers, err := monitoringProvider.GetMonitoringProvidersByAccountId(accountId)
		if err != nil {
			continue
		}
		for _, mp := range providers {
			if mp["monitoringProvider"] == mpType {
				return mp
			}
		}
	}
	return nil
}

// addVerifyWarning adds a standard Terraform warning diagnostic when a resource
// was found on the backend after a failed POST (lost response recovery).
func addVerifyWarning(resp *resource.CreateResponse, resourceType, name, id string) {
	resp.Diagnostics.AddWarning(
		resourceType+" created despite connection error",
		"'"+name+"' was found on the backend after a failed POST — "+
			"the response was likely lost in transit. Using existing ID: "+id,
	)
}
