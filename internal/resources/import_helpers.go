package resources

import "strings"

// splitImportID splits a composite import ID in the format "parent_id/child_id"
// and returns [parent_id, child_id].  Returns nil when the format is invalid
// (empty segments or missing separator).
func splitImportID(id string) []string {
	parts := strings.SplitN(id, "/", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return nil
	}
	return parts
}
