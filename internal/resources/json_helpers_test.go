package resources

import (
	"testing"
)

func TestReconcileJSONStateKeepsSemanticallyEqualPriorValue(t *testing.T) {
	prior := "{\n  \"b\": 2,\n  \"a\": [1, 2]\n}"
	backend := map[string]interface{}{"a": []interface{}{float64(1), float64(2)}, "b": float64(2)}

	got, err := reconcileJSONState(prior, backend)
	if err != nil {
		t.Fatalf("reconcileJSONState() error = %v", err)
	}
	if got != prior {
		t.Fatalf("reconcileJSONState() = %q, want prior value preserved", got)
	}
}

func TestReconcileJSONStateReturnsCanonicalOnDrift(t *testing.T) {
	prior := `{"a":1}`
	backend := map[string]interface{}{"a": float64(2)}

	got, err := reconcileJSONState(prior, backend)
	if err != nil {
		t.Fatalf("reconcileJSONState() error = %v", err)
	}
	if got != `{"a":2}` {
		t.Fatalf("reconcileJSONState() = %q, want canonical backend encoding", got)
	}
}

func TestReconcileJSONStateHandlesEmptyPrior(t *testing.T) {
	got, err := reconcileJSONState("", map[string]interface{}{})
	if err != nil {
		t.Fatalf("reconcileJSONState() error = %v", err)
	}
	if got != "{}" {
		t.Fatalf("reconcileJSONState() = %q, want {}", got)
	}
}

func TestParseJSONObjectRejectsNonObject(t *testing.T) {
	if _, err := parseJSONObject("definition", `[1,2]`); err == nil {
		t.Fatal("parseJSONObject() expected error for array input")
	}
	if _, err := parseJSONObject("definition", `not json`); err == nil {
		t.Fatal("parseJSONObject() expected error for invalid JSON")
	}
	out, err := parseJSONObject("definition", "")
	if err != nil || len(out) != 0 {
		t.Fatalf("parseJSONObject(\"\") = %v, %v; want empty map", out, err)
	}
}

func TestParseJSONArrayRejectsNonArray(t *testing.T) {
	if _, err := parseJSONArray("variable_schema", `{"a":1}`); err == nil {
		t.Fatal("parseJSONArray() expected error for object input")
	}
	out, err := parseJSONArray("variable_schema", "")
	if err != nil || len(out) != 0 {
		t.Fatalf("parseJSONArray(\"\") = %v, %v; want empty slice", out, err)
	}
}

func TestSupportedConfigTemplateTypes(t *testing.T) {
	for _, want := range []string{"naming_convention", "secret_pattern", "secret_naming_policy", "discovery_policy", "exclusion_rule", "resource_visibility_policy", "parser"} {
		if !isSupportedConfigTemplateType(want) {
			t.Fatalf("isSupportedConfigTemplateType(%q) = false", want)
		}
	}
	if isSupportedConfigTemplateType("config_file") {
		t.Fatal("isSupportedConfigTemplateType(config_file) = true; agent file templates are not Configuration Management templates")
	}
}

func TestProjectTemplateSubtypesExcludeCertificates(t *testing.T) {
	if _, ok := projectTemplateSubtypes["certificates"]; ok {
		t.Fatal("certificates category must be handled by mazevault_certificate_template")
	}
	for _, category := range []string{"secrets", "keys"} {
		subtypes := projectTemplateSubtypes[category]
		if len(subtypes) == 0 || subtypes[len(subtypes)-1] != "custom" {
			t.Fatalf("category %q must list subtypes ending with custom, got %v", category, subtypes)
		}
	}
}
