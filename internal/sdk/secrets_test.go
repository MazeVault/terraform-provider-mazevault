package mazevault

import (
	"encoding/json"
	"strings"
	"testing"
)

func TestUpdateSecretRequestOmitValueWhenNil(t *testing.T) {
	payload, err := json.Marshal(UpdateSecretRequest{
		TTLHours: 24,
		Metadata: map[string]string{"owner": "platform"},
	})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	encoded := string(payload)
	if strings.Contains(encoded, "value") {
		t.Fatalf("expected nil value to be omitted, got %s", encoded)
	}
	if !strings.Contains(encoded, "ttl_hours") || !strings.Contains(encoded, "metadata") {
		t.Fatalf("expected ttl_hours and metadata to be encoded, got %s", encoded)
	}
}

func TestUpdateSecretRequestKeepsExplicitEmptyValue(t *testing.T) {
	value := ""
	payload, err := json.Marshal(UpdateSecretRequest{Value: &value})
	if err != nil {
		t.Fatalf("marshal request: %v", err)
	}

	encoded := string(payload)
	if !strings.Contains(encoded, `"value":""`) {
		t.Fatalf("expected explicit empty value to be encoded, got %s", encoded)
	}
}
