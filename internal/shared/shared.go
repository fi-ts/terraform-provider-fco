package shared

import (
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/diag"

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
