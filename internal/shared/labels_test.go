package shared

import (
	"context"
	"testing"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
)

func mapValue(t *testing.T, m map[string]string) types.Map {
	t.Helper()
	elems := make(map[string]attr.Value, len(m))
	for k, v := range m {
		elems[k] = types.StringValue(v)
	}
	got, diags := types.MapValue(types.StringType, elems)
	if diags.HasError() {
		t.Fatalf("building map value: %v", diags)
	}
	return got
}

func TestLabelsToAPI(t *testing.T) {
	assert.Nil(t, LabelsToAPI(types.MapNull(types.StringType)))
	assert.Nil(t, LabelsToAPI(types.MapUnknown(types.StringType)))
	assert.Nil(t, LabelsToAPI(mapValue(t, map[string]string{})))

	got := LabelsToAPI(mapValue(t, map[string]string{"env": "prod"}))
	assert.NotNil(t, got)
	assert.Equal(t, map[string]string{"env": "prod"}, got.Labels)
}

func TestLabelsFromAPI(t *testing.T) {
	assert.True(t, LabelsFromAPI(nil).IsNull())
	assert.True(t, LabelsFromAPI(&apiv1.Labels{}).IsNull())

	got := LabelsFromAPI(&apiv1.Labels{Labels: map[string]string{"a": "1"}})
	assert.False(t, got.IsNull())
	var elems map[string]string
	assert.False(t, got.ElementsAs(context.Background(), &elems, false).HasError())
	assert.Equal(t, "1", elems["a"])
}

func TestLabelsToMap(t *testing.T) {
	assert.Equal(t, map[string]string{}, labelsToMap(types.MapNull(types.StringType)))
	assert.Equal(t, map[string]string{}, labelsToMap(types.MapUnknown(types.StringType)))
	assert.Equal(t, map[string]string{"a": "1"}, labelsToMap(mapValue(t, map[string]string{"a": "1"})))
}

func TestLabelsDelta(t *testing.T) {
	// No change -> nil.
	assert.Nil(t, LabelsDelta(types.MapNull(types.StringType), types.MapNull(types.StringType)))
	assert.Nil(t, LabelsDelta(mapValue(t, map[string]string{"a": "1"}), mapValue(t, map[string]string{"a": "1"})))

	// Add only.
	add := LabelsDelta(types.MapNull(types.StringType), mapValue(t, map[string]string{"a": "1"}))
	assert.NotNil(t, add)
	assert.Equal(t, map[string]string{"a": "1"}, add.Update.Labels)
	assert.Empty(t, add.Remove)

	// Remove only.
	rem := LabelsDelta(mapValue(t, map[string]string{"a": "1", "b": "2"}), mapValue(t, map[string]string{"b": "2"}))
	assert.NotNil(t, rem)
	assert.Nil(t, rem.Update)
	assert.Equal(t, []string{"a"}, rem.Remove)

	// Add + overwrite + remove.
	both := LabelsDelta(mapValue(t, map[string]string{"a": "1", "c": "3"}), mapValue(t, map[string]string{"a": "1", "b": "2", "d": "4"}))
	assert.NotNil(t, both)
	assert.Equal(t, map[string]string{"b": "2", "d": "4"}, both.Update.Labels)
	assert.Equal(t, []string{"c"}, both.Remove)

	// Changed value.
	changed := LabelsDelta(mapValue(t, map[string]string{"env": "dev"}), mapValue(t, map[string]string{"env": "prod", "owner": "fco"}))
	assert.NotNil(t, changed)
	assert.Equal(t, map[string]string{"env": "prod", "owner": "fco"}, changed.Update.Labels)

	// Plan null clears all state labels.
	clearAll := LabelsDelta(mapValue(t, map[string]string{"a": "1", "b": "2"}), types.MapNull(types.StringType))
	assert.NotNil(t, clearAll)
	assert.Nil(t, clearAll.Update)
	assert.ElementsMatch(t, []string{"a", "b"}, clearAll.Remove)
}
