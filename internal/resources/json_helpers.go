package resources

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
)

// parseJSONObject decodes a JSON object attribute (built with jsonencode()) into
// a generic map. Empty input yields an empty map so callers can treat omitted
// documents uniformly.
func parseJSONObject(attribute, raw string) (map[string]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return map[string]interface{}{}, nil
	}
	var out map[string]interface{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("%s must be a JSON object: %w", attribute, err)
	}
	if out == nil {
		out = map[string]interface{}{}
	}
	return out, nil
}

// parseJSONArray decodes a JSON array attribute into a generic slice. Empty input
// yields an empty slice.
func parseJSONArray(attribute, raw string) ([]interface{}, error) {
	if strings.TrimSpace(raw) == "" {
		return []interface{}{}, nil
	}
	var out []interface{}
	if err := json.Unmarshal([]byte(raw), &out); err != nil {
		return nil, fmt.Errorf("%s must be a JSON array: %w", attribute, err)
	}
	if out == nil {
		out = []interface{}{}
	}
	return out, nil
}

// reconcileJSONState returns the string that should be stored in state for a
// JSON document attribute: the prior (user-formatted) value when it is
// semantically identical to what the backend holds, otherwise the canonical
// backend encoding. This keeps key order and whitespace differences from
// surfacing as perpetual plan diffs while still detecting real drift.
func reconcileJSONState(prior string, backend interface{}) (string, error) {
	canonical, err := json.Marshal(backend)
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(prior) != "" {
		var priorValue interface{}
		if json.Unmarshal([]byte(prior), &priorValue) == nil {
			var backendValue interface{}
			if json.Unmarshal(canonical, &backendValue) == nil && reflect.DeepEqual(priorValue, backendValue) {
				return prior, nil
			}
		}
	}
	return string(canonical), nil
}
