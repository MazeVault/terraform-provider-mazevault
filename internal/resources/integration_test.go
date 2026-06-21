package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestIntegrationResource_MetadataAndSchema(t *testing.T) {
	r := NewIntegrationResource()
	if r == nil {
		t.Fatal("resource is nil")
	}
	var mdResp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "mazevault"}, &mdResp)
	if mdResp.TypeName != "mazevault_integration" {
		t.Fatalf("unexpected type name: %s", mdResp.TypeName)
	}
}
