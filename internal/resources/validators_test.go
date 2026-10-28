package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestNonNegativeInt64Validator(t *testing.T) {
	cases := []struct {
		name    string
		value   types.Int64
		wantErr bool
	}{
		{"positive", types.Int64Value(90), false},
		{"zero", types.Int64Value(0), false},
		{"negative", types.Int64Value(-1), true},
		{"null", types.Int64Null(), false},
		{"unknown", types.Int64Unknown(), false},
	}

	v := NonNegativeInt64()
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := validator.Int64Request{
				Path:        path.Root("rotation_interval_days"),
				ConfigValue: tc.value,
			}
			var resp validator.Int64Response
			v.ValidateInt64(context.Background(), req, &resp)
			if resp.Diagnostics.HasError() != tc.wantErr {
				t.Fatalf("wantErr=%v, got diagnostics: %v", tc.wantErr, resp.Diagnostics)
			}
		})
	}
}

func TestRotationTemplateSchemaAppliesNonNegativeValidators(t *testing.T) {
	r := NewRotationTemplateResource()
	var resp resource.SchemaResponse
	r.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}
	if diags := resp.Schema.ValidateImplementation(context.Background()); diags.HasError() {
		t.Fatalf("schema implementation diagnostics: %v", diags)
	}

	for _, field := range []string{
		"rotation_interval_days",
		"lead_time_days",
		"grace_period_days",
		"max_retry_attempts",
		"timeout_minutes",
	} {
		attr, ok := resp.Schema.Attributes[field].(resourceschema.Int64Attribute)
		if !ok {
			t.Fatalf("attribute %q is not an Int64Attribute", field)
		}
		if len(attr.Int64Validators()) == 0 {
			t.Fatalf("attribute %q must carry a non-negative validator", field)
		}
	}
}
