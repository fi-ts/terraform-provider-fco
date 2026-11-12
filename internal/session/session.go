package session

import (
	client "github.com/fi-ts/api/go/client"
)

// Session holds the authenticated API client and tenant context.
type Session struct {
	Client client.Client
	Tenant string
}
