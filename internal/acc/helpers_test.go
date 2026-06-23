package acc

import (
	"fmt"
	"os"
	"testing"

	"github.com/MazeVault/maze-core/terraform-provider-mazevault/internal/provider"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
)

// testAccProtoV6ProviderFactories wires up the provider in-process so that
// acceptance tests never need a built binary or dev_overrides.
var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"mazevault": providerserver.NewProtocol6WithError(provider.New("test")()),
}

// testAccPreCheck skips the test when required environment variables are absent.
// All acceptance tests must call t.Helper() + testAccPreCheck(t) as their
// first two statements.
func testAccPreCheck(t *testing.T) {
	t.Helper()
	if v := os.Getenv("MAZEVAULT_SERVER_URL"); v == "" {
		t.Skip("MAZEVAULT_SERVER_URL not set; skipping acceptance test")
	}
	if v := os.Getenv("MAZEVAULT_API_TOKEN"); v == "" {
		t.Skip("MAZEVAULT_API_TOKEN not set; skipping acceptance test")
	}
}

// providerConfig returns an HCL provider block that reads credentials from
// environment variables and enables skip_tls_verify for local / self-signed
// backends. Prepend to every test config.
func providerConfig() string {
	return fmt.Sprintf(`
provider "mazevault" {
  server_url      = %q
  api_token       = %q
  skip_tls_verify = true
}
`, os.Getenv("MAZEVAULT_SERVER_URL"), os.Getenv("MAZEVAULT_API_TOKEN"))
}
