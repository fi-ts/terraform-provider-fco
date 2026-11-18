package shared

import (
	"testing"
	"time"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestParseUpdatedAt(t *testing.T) {
	_, err := ParseUpdatedAt("")
	assert.ErrorIs(t, err, ErrNoUpdatedAt)

	_, err = ParseUpdatedAt("not-a-time")
	assert.Error(t, err)
	assert.NotErrorIs(t, err, ErrNoUpdatedAt)
}

func TestTimestampRoundTrip(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 30, 45, 123, time.UTC)
	formatted := TimestampFromAPI(timestamppb.New(now))
	assert.NotEmpty(t, formatted)

	ts, err := ParseUpdatedAt(formatted)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixNano(), ts.AsTime().UnixNano())
}

func TestTimestampFromAPIEmpty(t *testing.T) {
	assert.Equal(t, "", TimestampFromAPI(nil))
	assert.Equal(t, "", TimestampFromAPI((&apiv1.Meta{}).GetUpdatedAt()))
}
