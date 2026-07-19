package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
	resourceschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestSecretResourceSchemaValidateImplementation(t *testing.T) {
	secretResource := NewSecretResource()
	var resp resource.SchemaResponse
	secretResource.Schema(context.Background(), resource.SchemaRequest{}, &resp)
	if resp.Diagnostics.HasError() {
		t.Fatalf("schema diagnostics: %v", resp.Diagnostics)
	}

	diags := resp.Schema.ValidateImplementation(context.Background())
	if diags.HasError() {
		t.Fatalf("schema implementation diagnostics: %v", diags)
	}

	attr, ok := resp.Schema.Attributes["value_wo"].(resourceschema.StringAttribute)
	if !ok {
		t.Fatalf("value_wo attribute type = %T, want schema.StringAttribute", resp.Schema.Attributes["value_wo"])
	}
	if !attr.WriteOnly || !attr.Optional || attr.Computed {
		t.Fatalf("value_wo schema flags = write_only:%v optional:%v computed:%v", attr.WriteOnly, attr.Optional, attr.Computed)
	}
}

func TestShouldSendWriteOnlyValue(t *testing.T) {
	tests := []struct {
		name     string
		previous types.Int64
		next     types.Int64
		want     bool
	}{
		{
			name:     "first configured version",
			previous: types.Int64Null(),
			next:     types.Int64Value(1),
			want:     true,
		},
		{
			name:     "unchanged version",
			previous: types.Int64Value(1),
			next:     types.Int64Value(1),
			want:     false,
		},
		{
			name:     "changed version",
			previous: types.Int64Value(1),
			next:     types.Int64Value(2),
			want:     true,
		},
		{
			name:     "missing planned version",
			previous: types.Int64Value(1),
			next:     types.Int64Null(),
			want:     false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := shouldSendWriteOnlyValue(tt.previous, tt.next)
			if got != tt.want {
				t.Fatalf("shouldSendWriteOnlyValue() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestSourceCount(t *testing.T) {
	if got := sourceCount(true, false, true); got != 2 {
		t.Fatalf("sourceCount() = %d, want 2", got)
	}
	if got := sourceCount(false, false, false); got != 0 {
		t.Fatalf("sourceCount() = %d, want 0", got)
	}
}
