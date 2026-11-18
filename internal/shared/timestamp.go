package shared

import (
	"errors"
	"fmt"
	"time"

	"google.golang.org/protobuf/types/known/timestamppb"
)

// ErrNoUpdatedAt is returned by ParseUpdatedAt when the stored value is empty
// (no timestamp available for optimistic locking). Callers can distinguish
// this from a genuine parse error with errors.Is.
var ErrNoUpdatedAt = errors.New("no updated_at timestamp available for optimistic locking")

// TimestampFromAPI formats a protobuf Timestamp (possibly nil) as RFC 3339
// Nano, or "" when absent.
func TimestampFromAPI(ts *timestamppb.Timestamp) string {
	if ts == nil {
		return ""
	}
	return ts.AsTime().Format(time.RFC3339Nano)
}

// ParseUpdatedAt parses an RFC 3339 / RFC 3339 Nano timestamp back into a
// protobuf Timestamp.
func ParseUpdatedAt(s string) (*timestamppb.Timestamp, error) {
	if s == "" {
		return nil, ErrNoUpdatedAt
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return nil, fmt.Errorf("parse updated_at %q: %w", s, err)
	}
	return timestamppb.New(t), nil
}
