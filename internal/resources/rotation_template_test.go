package resources

import (
	"testing"

	mazevault "github.com/MazeVault/maze-core/sdks/go"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestFlattenRotationTemplateMapsAllFields(t *testing.T) {
	tmpl := &mazevault.RotationConfigTemplate{
		ID:                   "tmpl-1",
		Name:                 "db-monthly",
		Description:          "desc",
		OrganizationID:       "org-1",
		ProjectID:            "proj-1",
		IsDefault:            true,
		RotationIntervalDays: 30,
		LeadTimeDays:         7,
		GracePeriodDays:      2,
		MaxRetryAttempts:     4,
		RetryDelaySeconds:    15,
		TimeoutMinutes:       20,
		SecretCustodyMode:    "transient",
	}
	var data RotationTemplateModel
	flattenRotationTemplate(tmpl, &data)

	if data.ID.ValueString() != "tmpl-1" || data.OrganizationID.ValueString() != "org-1" {
		t.Fatalf("id/org mismatch: %+v", data)
	}
	if data.ProjectID.ValueString() != "proj-1" {
		t.Fatalf("project_id = %q, want proj-1", data.ProjectID.ValueString())
	}
	if data.RetryDelaySeconds.ValueInt64() != 15 {
		t.Fatalf("retry_delay_seconds = %d, want 15", data.RetryDelaySeconds.ValueInt64())
	}
	if !data.IsDefault.ValueBool() || data.SecretCustodyMode.ValueString() != "transient" {
		t.Fatalf("is_default/custody mismatch: %+v", data)
	}
}

func TestFlattenRotationTemplateEmptyProjectIsNull(t *testing.T) {
	var data RotationTemplateModel
	flattenRotationTemplate(&mazevault.RotationConfigTemplate{ID: "t", ProjectID: ""}, &data)
	if !data.ProjectID.IsNull() {
		t.Fatalf("empty project_id must map to null, got %q", data.ProjectID.ValueString())
	}
}

func TestBuildRotationTemplateRequestIncludesNewFields(t *testing.T) {
	data := RotationTemplateModel{
		Name:              types.StringValue("db"),
		ProjectID:         types.StringValue("proj-1"),
		RetryDelaySeconds: types.Int64Value(11),
	}
	req := buildRotationTemplateRequest(data)
	if req.ProjectID != "proj-1" {
		t.Fatalf("project_id = %q, want proj-1", req.ProjectID)
	}
	if req.RetryDelaySeconds != 11 {
		t.Fatalf("retry_delay_seconds = %d, want 11", req.RetryDelaySeconds)
	}
}
