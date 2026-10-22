package session

import (
	client "github.com/fi-ts/api/go/client"
)

// Session holds the authenticated API client and project/tenant context.
type Session struct {
	Client      client.Client
	ProjectUUID string
	Tenant      string
}
