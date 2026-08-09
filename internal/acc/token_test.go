package acc

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/terraform"
)

func TestAccTokenResource_writeOnlyValue(t *testing.T) {
	t.Helper()

	const tokenID = "11111111-1111-4111-8111-111111111111"
	const writeOnlyValue = "acceptance-test-token-value"
	var mutex sync.Mutex
	description := "initial"
	created := false
	createValue := ""
	readStatus := 0

	server := httptest.NewServer(http.HandlerFunc(func(response http.ResponseWriter, request *http.Request) {
		mutex.Lock()
		defer mutex.Unlock()

		response.Header().Set("Content-Type", "application/json")
		switch {
		case request.Method == http.MethodPost && request.URL.Path == "/api/v1/tokens":
			var body struct {
				Description string `json:"description"`
				Value       string `json:"value"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			createValue = body.Value
			description = body.Description
			created = true
			writeTokenResponse(t, response, tokenID, description)
		case request.Method == http.MethodGet && request.URL.Path == "/api/v1/tokens/"+tokenID:
			if readStatus != 0 {
				http.Error(response, `{"error":"temporary failure"}`, readStatus)
				return
			}
			if !created {
				http.Error(response, `{"error":"not found"}`, http.StatusNotFound)
				return
			}
			writeTokenResponse(t, response, tokenID, description)
		case request.Method == http.MethodPut && request.URL.Path == "/api/v1/tokens/"+tokenID:
			var body struct {
				Description *string `json:"description"`
				Value       *string `json:"value"`
			}
			if err := json.NewDecoder(request.Body).Decode(&body); err != nil {
				http.Error(response, err.Error(), http.StatusBadRequest)
				return
			}
			if body.Value != nil {
				t.Errorf("metadata update unexpectedly resent token value")
			}
			if body.Description != nil {
				description = *body.Description
			}
			writeTokenResponse(t, response, tokenID, description)
		case request.Method == http.MethodDelete && request.URL.Path == "/api/v1/tokens/"+tokenID:
			created = false
			http.Error(response, `{"error":"not found"}`, http.StatusNotFound)
		default:
			http.Error(response, fmt.Sprintf("unexpected request: %s %s", request.Method, request.URL.Path), http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := func(currentDescription string) string {
		return fmt.Sprintf(`
provider "mazevault" {
  server_url = %q
  api_token  = "test-auth-token"
}

resource "mazevault_token" "test" {
  name             = "terraform-acceptance-token"
  token_provider   = "generic"
  description      = %q
  value_wo         = %q
  value_wo_version = 1
  expires_at       = "2027-01-01T00:00:00Z"
}
`, server.URL, currentDescription, writeOnlyValue)
	}

	resource.Test(t, resource.TestCase{
		ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
		Steps: []resource.TestStep{
			{
				Config: config("initial"),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_token.test", "id", tokenID),
					resource.TestCheckResourceAttr("mazevault_token.test", "token_provider", "generic"),
					resource.TestCheckResourceAttr("mazevault_token.test", "description", "initial"),
					resource.TestCheckResourceAttr("mazevault_token.test", "value_wo_version", "1"),
					resource.TestCheckNoResourceAttr("mazevault_token.test", "value_wo"),
					func(_ *terraform.State) error {
						mutex.Lock()
						defer mutex.Unlock()
						if createValue != writeOnlyValue {
							return fmt.Errorf("create value was not delivered through value_wo")
						}
						return nil
					},
				),
			},
			{
				Config: config(""),
				Check: resource.ComposeTestCheckFunc(
					resource.TestCheckResourceAttr("mazevault_token.test", "description", ""),
					resource.TestCheckNoResourceAttr("mazevault_token.test", "value_wo"),
				),
			},
			{
				ResourceName:            "mazevault_token.test",
				ImportState:             true,
				ImportStateVerify:       true,
				ImportStateVerifyIgnore: []string{"description", "value_wo", "value_wo_version"},
			},
			{
				PreConfig: func() {
					mutex.Lock()
					readStatus = http.StatusInternalServerError
					mutex.Unlock()
				},
				Config:      config(""),
				PlanOnly:    true,
				ExpectError: regexp.MustCompile("Read Token Error"),
			},
			{
				PreConfig: func() {
					mutex.Lock()
					readStatus = 0
					created = false
					mutex.Unlock()
				},
				Config:             config(""),
				PlanOnly:           true,
				ExpectNonEmptyPlan: true,
			},
		},
	})
}

func writeTokenResponse(t *testing.T, response http.ResponseWriter, id, description string) {
	t.Helper()
	now := time.Date(2026, time.August, 9, 12, 0, 0, 0, time.UTC)
	payload := map[string]interface{}{
		"id":                     id,
		"organization_id":        "22222222-2222-4222-8222-222222222222",
		"name":                   "terraform-acceptance-token",
		"description":            description,
		"provider":               "generic",
		"storage_mode":           "local",
		"expires_at":             "2027-01-01T00:00:00Z",
		"lead_time_days":         30,
		"status":                 "active",
		"is_expired":             false,
		"days_until_expiry":      145,
		"created_at":             now.Format(time.RFC3339),
		"updated_at":             now.Format(time.RFC3339),
		"deployment_targets":     []interface{}{},
		"external_provider_type": "",
	}
	if err := json.NewEncoder(response).Encode(payload); err != nil {
		t.Fatalf("encode token response: %v", err)
	}
}
