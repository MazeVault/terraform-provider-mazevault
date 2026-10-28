package mazevault

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestCreateRotationTemplateRequestEncodesNewFields(t *testing.T) {
	payload, err := json.Marshal(CreateRotationTemplateRequest{
		Name:              "db-monthly",
		ProjectID:         "11111111-1111-1111-1111-111111111111",
		RetryDelaySeconds: 15,
		SecretCustodyMode: "transient",
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}
	encoded := string(payload)
	for _, want := range []string{
		`"project_id":"11111111-1111-1111-1111-111111111111"`,
		`"retry_delay_seconds":15`,
		`"secret_custody_mode":"transient"`,
	} {
		if !strings.Contains(encoded, want) {
			t.Fatalf("expected %s in %s", want, encoded)
		}
	}
}

func TestRotationConfigTemplateDecodesNewFields(t *testing.T) {
	var tmpl RotationConfigTemplate
	body := `{"id":"t1","name":"db","project_id":"p1","retry_delay_seconds":9,"secret_custody_mode":"delegated"}`
	if err := json.Unmarshal([]byte(body), &tmpl); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if tmpl.ProjectID != "p1" {
		t.Fatalf("project_id = %q, want p1", tmpl.ProjectID)
	}
	if tmpl.RetryDelaySeconds != 9 {
		t.Fatalf("retry_delay_seconds = %d, want 9", tmpl.RetryDelaySeconds)
	}
	if tmpl.SecretCustodyMode != "delegated" {
		t.Fatalf("secret_custody_mode = %q, want delegated", tmpl.SecretCustodyMode)
	}
}
