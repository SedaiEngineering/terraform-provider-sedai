package provider

import "github.com/hashicorp/terraform-plugin-framework/types/basetypes"

// This file is the single home for `basetypes.X` ↔ `*T` pointer
// conversions. Every per-resource-type settings block in Stage 9 (kube,
// serverless, bucket, volume, ECS, container, app) has many optional
// fields that translate the same way; centralizing here avoids 100+ lines
// of duplicated null-checks across blocks.
//
// Convention: the *Ptr helpers return nil for null/unknown TF values
// (meaning "user didn't manage this field"). The nullable* helpers go
// the other direction — converting an SDK-side *T into a TF state value,
// using `New<Type>Null()` when the pointer is nil.

// stringPtr converts a TF StringValue into a *string. Returns nil for
// null/unknown values so downstream SDK code knows to leave the backend's
// value untouched.
func stringPtr(v basetypes.StringValue) *string {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	s := v.ValueString()
	return &s
}

// boolPtr converts a TF BoolValue into a *bool.
func boolPtr(v basetypes.BoolValue) *bool {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	b := v.ValueBool()
	return &b
}

// int64Ptr converts a TF Int64Value into a *int64.
func int64Ptr(v basetypes.Int64Value) *int64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	i := v.ValueInt64()
	return &i
}

// float64Ptr converts a TF Float64Value into a *float64.
func float64Ptr(v basetypes.Float64Value) *float64 {
	if v.IsNull() || v.IsUnknown() {
		return nil
	}
	f := v.ValueFloat64()
	return &f
}

// nullableString returns a TF StringValue — known when v is non-nil, null
// when v is nil. Used to populate computed / refreshed state fields from
// SDK-side optional values.
func nullableString(v *string) basetypes.StringValue {
	if v == nil {
		return basetypes.NewStringNull()
	}
	return basetypes.NewStringValue(*v)
}

// nullableBool — see nullableString.
func nullableBool(v *bool) basetypes.BoolValue {
	if v == nil {
		return basetypes.NewBoolNull()
	}
	return basetypes.NewBoolValue(*v)
}

// nullableInt64 — see nullableString.
func nullableInt64(v *int64) basetypes.Int64Value {
	if v == nil {
		return basetypes.NewInt64Null()
	}
	return basetypes.NewInt64Value(*v)
}

// nullableFloat64 — see nullableString.
func nullableFloat64(v *float64) basetypes.Float64Value {
	if v == nil {
		return basetypes.NewFloat64Null()
	}
	return basetypes.NewFloat64Value(*v)
}

// The refresh* helpers below implement the field-level partial-spec
// contract on the Read path: a tfsdk field that's already null/unknown
// in state means "user is not managing this field", so we leave it null
// rather than overwriting with the backend's value. A field that's
// already known means "user manages it", so we refresh from the backend
// — that's how UI-side drift gets detected.
//
// Used by every per-resource-type settings block (kube, bucket, etc.)
// so they live here instead of duplicated per-block.

// refreshIfManaged updates a StringValue from a *string IF state already
// had a known value. Null/unknown state stays null (unmanaged).
func refreshIfManaged(state *basetypes.StringValue, fetched *string) {
	if state.IsNull() || state.IsUnknown() {
		return
	}
	*state = nullableString(fetched)
}

// refreshBoolIfManaged — see refreshIfManaged.
func refreshBoolIfManaged(state *basetypes.BoolValue, fetched *bool) {
	if state.IsNull() || state.IsUnknown() {
		return
	}
	*state = nullableBool(fetched)
}

// refreshInt64IfManaged — see refreshIfManaged.
func refreshInt64IfManaged(state *basetypes.Int64Value, fetched *int64) {
	if state.IsNull() || state.IsUnknown() {
		return
	}
	*state = nullableInt64(fetched)
}

// refreshFloat64IfManaged — see refreshIfManaged.
func refreshFloat64IfManaged(state *basetypes.Float64Value, fetched *float64) {
	if state.IsNull() || state.IsUnknown() {
		return
	}
	*state = nullableFloat64(fetched)
}
