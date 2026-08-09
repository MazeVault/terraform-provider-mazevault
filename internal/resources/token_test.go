package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestTokenResourceSchemaValidateImplementation(t *testing.T) {
	tokenResource := NewTokenResource()
	var response resource.SchemaResponse
	tokenResource.Schema(context.Background(), resource.SchemaRequest{}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", response.Diagnostics)
	}

	diagnostics := response.Schema.ValidateImplementation(context.Background())
	if diagnostics.HasError() {
		t.Fatalf("schema implementation diagnostics: %v", diagnostics)
	}

	writeOnly, ok := response.Schema.Attributes["value_wo"].(resourceschema.StringAttribute)
	if !ok {
		t.Fatalf("value_wo attribute type = %T, want schema.StringAttribute", response.Schema.Attributes["value_wo"])
	}
	if !writeOnly.WriteOnly || !writeOnly.Optional || writeOnly.Computed || len(writeOnly.PlanModifiers) != 0 {
		t.Fatalf("value_wo schema flags = write_only:%v optional:%v computed:%v modifiers:%d", writeOnly.WriteOnly, writeOnly.Optional, writeOnly.Computed, len(writeOnly.PlanModifiers))
	}

	version, ok := response.Schema.Attributes["value_wo_version"].(resourceschema.Int64Attribute)
	if !ok || len(version.PlanModifiers) != 1 {
		t.Fatalf("value_wo_version must have one replacement plan modifier")
	}
	if _, exists := response.Schema.Attributes["provider"]; exists {
		t.Fatal("provider is a reserved root attribute name")
	}
	if _, exists := response.Schema.Attributes["token_provider"]; !exists {
		t.Fatal("token_provider attribute is missing")
	}
	if response.Schema.Version != 1 {
		t.Fatalf("schema version = %d, want 1", response.Schema.Version)
	}
}

func TestTokenResourceUpgradeStateV0(t *testing.T) {
	ctx := context.Background()
	prior := tokenModelV0{
		ID:                   types.StringValue("11111111-1111-4111-8111-111111111111"),
		Name:                 types.StringValue("legacy-token"),
		Description:          types.StringValue("legacy description"),
		Provider:             types.StringValue("jira"),
		Value:                types.StringValue("legacy-state-value"),
		ProjectID:            types.StringNull(),
		ExpiresAt:            types.StringValue("2027-01-01T00:00:00Z"),
		LeadTimeDays:         types.Int64Value(30),
		Status:               types.StringValue("active"),
		StorageMode:          types.StringValue("local"),
		ExternalProviderType: types.StringNull(),
		IsExpired:            types.BoolValue(false),
		DaysUntilExpiry:      types.Int64Value(145),
		CreatedAt:            types.StringValue("2026-08-09T12:00:00Z"),
		UpdatedAt:            types.StringValue("2026-08-09T12:00:00Z"),
	}
	priorState := tfsdk.State{Schema: tokenSchemaV0()}
	if diagnostics := priorState.Set(ctx, &prior); diagnostics.HasError() {
		t.Fatalf("set prior state: %v", diagnostics)
	}

	tokenResource := &TokenResource{}
	var schemaResponse resource.SchemaResponse
	tokenResource.Schema(ctx, resource.SchemaRequest{}, &schemaResponse)
	response := resource.UpgradeStateResponse{State: tfsdk.State{Schema: schemaResponse.Schema}}
	upgrader := tokenResource.UpgradeState(ctx)[0]
	upgrader.StateUpgrader(ctx, resource.UpgradeStateRequest{State: &priorState}, &response)
	if response.Diagnostics.HasError() {
		t.Fatalf("upgrade diagnostics: %v", response.Diagnostics)
	}

	var upgraded TokenModel
	if diagnostics := response.State.Get(ctx, &upgraded); diagnostics.HasError() {
		t.Fatalf("get upgraded state: %v", diagnostics)
	}
	if upgraded.TokenProvider.ValueString() != "jira" {
		t.Fatalf("token_provider = %q, want jira", upgraded.TokenProvider.ValueString())
	}
	if upgraded.Value.ValueString() != "legacy-state-value" {
		t.Fatal("legacy value was not preserved during state upgrade")
	}
	if !upgraded.ValueWO.IsNull() || !upgraded.ValueWOVersion.IsNull() {
		t.Fatal("new write-only fields must be null after v0 state upgrade")
	}
}

func TestReconcileOptionalString(t *testing.T) {
	tests := []struct {
		name     string
		apiValue string
		current  types.String
		want     types.String
	}{
		{name: "api value is authoritative", apiValue: "new", current: types.StringValue("old"), want: types.StringValue("new")},
		{name: "empty API clears stale value", current: types.StringValue("old"), want: types.StringNull()},
		{name: "explicit empty remains empty", current: types.StringValue(""), want: types.StringValue("")},
		{name: "null remains null", current: types.StringNull(), want: types.StringNull()},
		{name: "unknown resolves to null", current: types.StringUnknown(), want: types.StringNull()},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := reconcileOptionalString(test.apiValue, test.current)
			if !got.Equal(test.want) {
				t.Fatalf("reconcileOptionalString() = %#v, want %#v", got, test.want)
			}
		})
	}
}

func TestValidateTokenValueSources(t *testing.T) {
	tests := []struct {
		name         string
		legacy       types.String
		writeOnly    types.String
		version      types.Int64
		requireValue bool
		wantError    bool
	}{
		{name: "create accepts legacy", legacy: types.StringValue("legacy"), writeOnly: types.StringNull(), version: types.Int64Null(), requireValue: true},
		{name: "create accepts write only with version", legacy: types.StringNull(), writeOnly: types.StringValue("secure"), version: types.Int64Value(1), requireValue: true},
		{name: "create requires a value", legacy: types.StringNull(), writeOnly: types.StringNull(), version: types.Int64Null(), requireValue: true, wantError: true},
		{name: "metadata update accepts no value", legacy: types.StringNull(), writeOnly: types.StringNull(), version: types.Int64Null()},
		{name: "dual sources rejected", legacy: types.StringValue("legacy"), writeOnly: types.StringValue("secure"), version: types.Int64Value(1), wantError: true},
		{name: "write only requires version", legacy: types.StringNull(), writeOnly: types.StringValue("secure"), version: types.Int64Null(), wantError: true},
		{name: "orphan version rejected", legacy: types.StringNull(), writeOnly: types.StringNull(), version: types.Int64Value(1), wantError: true},
		{name: "unknown input rejected", legacy: types.StringUnknown(), writeOnly: types.StringNull(), version: types.Int64Null(), wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			message := validateTokenValueSources(test.legacy, test.writeOnly, test.version, test.requireValue)
			if (message != "") != test.wantError {
				t.Fatalf("validateTokenValueSources() error = %q, wantError = %v", message, test.wantError)
			}
		})
	}
}
