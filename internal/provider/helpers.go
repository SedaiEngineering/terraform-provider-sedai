package provider

import (
	"time"

	"github.com/SedaiEngineering/sedai-sdk-go/sdk/sedai/monitoringProvider"
	"github.com/hashicorp/terraform-plugin-framework/resource"
)

// safeMapString safely extracts a string value from a map[string]interface{}.
// Returns ("", false) if the key is missing or the value is not a non-empty string,
// instead of panicking on a nil or wrong-type assertion.
func safeMapString(m map[string]interface{}, key string) (string, bool) {
	if m == nil {
		return "", false
	}
	v, ok := m[key]
	if !ok || v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok && s != ""
}

// deleteMPGracefully deletes a monitoring provider and handles the case where
// the backend returns an error from a non-fatal cleanup step (e.g. exporter
// deregistration failure). If the delete call errors but the provider no longer
// exists on a subsequent Read, the delete is treated as successful.
//
// This is a provider-side workaround for a backend issue where the exporter
// deregistration step fails and causes the delete response to return an error
// even though the MP was actually removed. The correct long-term fix is to make
// exporter deregistration best-effort in sedai-core's delete handler.
func deleteMPGracefully(id string) error {
	_, err := monitoringProvider.DeleteMonitoringProvider(id)
	if err != nil {
		// Check if the resource was actually deleted despite the error.
		// If it's gone, suppress the error — delete succeeded.
		existing, fetchErr := monitoringProvider.GetMonitoringProviderById(id)
		if fetchErr != nil || existing == nil {
			return nil
		}
		return err
	}
	return nil
}

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
