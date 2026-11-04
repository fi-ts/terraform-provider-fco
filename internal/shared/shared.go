package shared

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/fi-ts/terraform-provider-fco/internal/session"
)

// Debug enables verbose provider logging.
var Debug bool

// SessionFrom extracts the *session.Session from the provider data.
func SessionFrom(providerData any) (*session.Session, diag.Diagnostic) {
	if providerData == nil {
		return nil, diag.NewErrorDiagnostic(
			"Missing FCO provider data",
			"The provider data was not set. Verify that the provider's Configure method ran successfully. This is always a provider bug.",
		)
	}
	s, ok := providerData.(*session.Session)
	if !ok {
		return nil, diag.NewErrorDiagnostic(
			"Unexpected provider data type",
			fmt.Sprintf("Expected *session.Session, got: %T. Please report this issue to the provider developers.", providerData),
		)
	}
	return s, nil
}

// ListID returns a stable content hash over the listed entity UUIDs, used as
// a data source's `id` so Terraform can detect when the listed set changes.
func ListID(uuids []string) types.String {
	return types.StringValue(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(uuids, ",")))))
}
