package resources

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/resource"
)

func TestSecretLinkResource_MetadataAndSchema(t *testing.T) {
	r := NewSecretLinkResource()
	if r == nil {
		t.Fatal("resource is nil")
	}
	// Metadata
	var mdResp resource.MetadataResponse
	r.Metadata(context.Background(), resource.MetadataRequest{ProviderTypeName: "mazevault"}, &mdResp)
	if mdResp.TypeName != "mazevault_secret_link" {
		t.Fatalf("unexpected type name: %s", mdResp.TypeName)
	}
}
