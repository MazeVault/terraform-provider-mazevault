package resources

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/schema/validator"
)

// nonNegativeInt64Validator rejects negative values for an Int64 attribute at plan
// time, mirroring the backend's non-negative policy for rotation template fields.
// Null/unknown values pass (the attribute is optional and may be computed server-side).
type nonNegativeInt64Validator struct{}

// NonNegativeInt64 returns a validator that ensures an Int64 attribute is >= 0.
func NonNegativeInt64() validator.Int64 { return nonNegativeInt64Validator{} }

func (v nonNegativeInt64Validator) Description(_ context.Context) string {
	return "value must be greater than or equal to 0"
}

func (v nonNegativeInt64Validator) MarkdownDescription(_ context.Context) string {
	return "value must be greater than or equal to `0`"
}

func (v nonNegativeInt64Validator) ValidateInt64(_ context.Context, req validator.Int64Request, resp *validator.Int64Response) {
	if req.ConfigValue.IsNull() || req.ConfigValue.IsUnknown() {
		return
	}
	if value := req.ConfigValue.ValueInt64(); value < 0 {
		resp.Diagnostics.AddAttributeError(
			req.Path,
			"Invalid value",
			fmt.Sprintf("%s must be greater than or equal to 0, got %d.", req.Path, value),
		)
	}
}
