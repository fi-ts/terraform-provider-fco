package shared

import (
	"crypto/sha1"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

// UUIDer is implemented by API objects that expose a UUID.
type UUIDer interface {
	GetUuid() string
}

// ListViews converts API list items into their Terraform views and derives a
// stable data source ID from the items' UUIDs.
func ListViews[T UUIDer, V any](items []T, view func(T) V) ([]V, types.String) {
	views := make([]V, 0, len(items))
	ids := make([]string, 0, len(items))
	for _, item := range items {
		views = append(views, view(item))
		ids = append(ids, item.GetUuid())
	}
	return views, types.StringValue(fmt.Sprintf("%x", sha1.Sum([]byte(strings.Join(ids, ",")))))
}
