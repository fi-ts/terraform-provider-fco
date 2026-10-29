package publiccloudvm

import (
	"context"
	"testing"
	"time"

	pcvmv1 "github.com/fi-ts/api/go/fits/api/pcvm/v1"
	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestSizeFromString(t *testing.T) {
	cases := []struct {
		in     string
		want   pcvmv1.Size
		wantOK bool
	}{
		{"small", pcvmv1.Size_SIZE_SMALL, true},
		{"medium", pcvmv1.Size_SIZE_MEDIUM, true},
		{"large", pcvmv1.Size_SIZE_LARGE, true},
		{"", pcvmv1.Size_SIZE_UNSPECIFIED, false},
		{"xl", pcvmv1.Size_SIZE_UNSPECIFIED, false},
		{"SMALL", pcvmv1.Size_SIZE_UNSPECIFIED, false},
	}
	for _, c := range cases {
		got, ok := sizeFromString(c.in)
		assert.Equal(t, c.wantOK, ok, "sizeFromString(%q) ok", c.in)
		if ok {
			assert.Equal(t, c.want, got, "sizeFromString(%q)", c.in)
		}
	}
}

func TestSizeToString(t *testing.T) {
	assert.Equal(t, "small", sizeToString(pcvmv1.Size_SIZE_SMALL))
	assert.Equal(t, "medium", sizeToString(pcvmv1.Size_SIZE_MEDIUM))
	assert.Equal(t, "large", sizeToString(pcvmv1.Size_SIZE_LARGE))
	assert.Equal(t, "unspecified", sizeToString(pcvmv1.Size_SIZE_UNSPECIFIED))
}

func TestStateToString(t *testing.T) {
	assert.Equal(t, "healthy", stateToString(pcvmv1.State_STATE_HEALTHY))
	assert.Equal(t, "warning", stateToString(pcvmv1.State_STATE_WARNING))
	assert.Equal(t, "error", stateToString(pcvmv1.State_STATE_ERROR))
	assert.Equal(t, "unspecified", stateToString(pcvmv1.State_STATE_UNSPECIFIED))
}

func TestOperationToString(t *testing.T) {
	assert.Equal(t, "creating", operationToString(pcvmv1.Operation_OPERATION_CREATING))
	assert.Equal(t, "active", operationToString(pcvmv1.Operation_OPERATION_ACTIVE))
	assert.Equal(t, "updating", operationToString(pcvmv1.Operation_OPERATION_UPDATING))
	assert.Equal(t, "deleted", operationToString(pcvmv1.Operation_OPERATION_DELETED))
	assert.Equal(t, "error", operationToString(pcvmv1.Operation_OPERATION_ERROR))
	assert.Equal(t, "unspecified", operationToString(pcvmv1.Operation_OPERATION_UNSPECIFIED))
}

func TestIsSizeDowngrade(t *testing.T) {
	assert.True(t, isSizeDowngrade(pcvmv1.Size_SIZE_LARGE, pcvmv1.Size_SIZE_MEDIUM))
	assert.True(t, isSizeDowngrade(pcvmv1.Size_SIZE_MEDIUM, pcvmv1.Size_SIZE_SMALL))
	assert.True(t, isSizeDowngrade(pcvmv1.Size_SIZE_LARGE, pcvmv1.Size_SIZE_SMALL))
	assert.False(t, isSizeDowngrade(pcvmv1.Size_SIZE_SMALL, pcvmv1.Size_SIZE_MEDIUM))
	assert.False(t, isSizeDowngrade(pcvmv1.Size_SIZE_MEDIUM, pcvmv1.Size_SIZE_LARGE))
	// Equal sizes are not a downgrade (upgrades-or-equal are allowed).
	assert.False(t, isSizeDowngrade(pcvmv1.Size_SIZE_SMALL, pcvmv1.Size_SIZE_SMALL))
	assert.False(t, isSizeDowngrade(pcvmv1.Size_SIZE_LARGE, pcvmv1.Size_SIZE_LARGE))
}

func TestLabelsToAPI(t *testing.T) {
	// Null map -> nil.
	assert.Nil(t, labelsToAPI(types.MapNull(types.StringType)))
	// Unknown map -> nil.
	assert.Nil(t, labelsToAPI(types.MapUnknown(types.StringType)))
	// Populated map -> Labels.
	one, _ := types.MapValue(types.StringType, map[string]attr.Value{"env": types.StringValue("prod")})
	got := labelsToAPI(one)
	assert.NotNil(t, got)
	assert.Equal(t, map[string]string{"env": "prod"}, got.Labels)
}

func TestLabelsFromAPI(t *testing.T) {
	// nil -> null.
	assert.True(t, labelsFromAPI(nil).IsNull())
	// empty -> null.
	assert.True(t, labelsFromAPI(&apiv1.Labels{}).IsNull())
	// populated -> value map.
	got := labelsFromAPI(&apiv1.Labels{Labels: map[string]string{"a": "1"}})
	assert.False(t, got.IsNull())
	var gotElems map[string]string
	assert.False(t, got.ElementsAs(context.Background(), &gotElems, false).HasError())
	assert.Equal(t, "1", gotElems["a"])
}

func TestUpdatedAtRoundTrip(t *testing.T) {
	now := time.Date(2024, 6, 15, 12, 30, 45, 0, time.UTC)
	meta := &apiv1.Meta{UpdatedAt: timestamppb.New(now)}
	formatted := updatedAtFromAPI(meta)
	assert.NotEmpty(t, formatted)

	ts, err := parseUpdatedAt(formatted)
	assert.NoError(t, err)
	assert.Equal(t, now.UnixNano(), ts.AsTime().UnixNano())
}

func TestUpdatedAtFromAPIEmpty(t *testing.T) {
	assert.Equal(t, "", updatedAtFromAPI(nil))
	assert.Equal(t, "", updatedAtFromAPI(&apiv1.Meta{}))
}

func TestParseUpdatedAtErrors(t *testing.T) {
	_, err := parseUpdatedAt("")
	assert.Error(t, err)
	_, err = parseUpdatedAt("not-a-time")
	assert.Error(t, err)
}

func TestPublicCloudVMFromAPI(t *testing.T) {
	vm := &pcvmv1.PublicCloudVM{
		Uuid:        "11111111-2222-3333-4444-555555555555",
		Name:        "my-vm",
		Size:        pcvmv1.Size_SIZE_MEDIUM,
		Tenant:      "tenant-abc",
		ProjectUuid: "99999999-8888-7777-6666-555555555555",
		Meta: &apiv1.Meta{
			Labels:    &apiv1.Labels{Labels: map[string]string{"team": "platform"}},
			UpdatedAt: timestamppb.New(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)),
		},
		Status: &pcvmv1.Status{
			State:     pcvmv1.State_STATE_HEALTHY,
			Operation: pcvmv1.Operation_OPERATION_ACTIVE,
			Info:      "all good",
		},
	}
	m := publicCloudVMFromAPI(vm)
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", m.UUID.ValueString())
	assert.Equal(t, "my-vm", m.Name.ValueString())
	assert.Equal(t, "medium", m.Size.ValueString())
	assert.Equal(t, "tenant-abc", m.Tenant.ValueString())
	assert.Equal(t, "99999999-8888-7777-6666-555555555555", m.ProjectUUID.ValueString())
	assert.Equal(t, "healthy", m.StatusState.ValueString())
	assert.Equal(t, "active", m.StatusOperation.ValueString())
	assert.Equal(t, "all good", m.StatusInfo.ValueString())
	assert.False(t, m.Labels.IsNull())
	var labelElems map[string]string
	assert.False(t, m.Labels.ElementsAs(context.Background(), &labelElems, false).HasError())
	assert.Equal(t, "platform", labelElems["team"])
	assert.NotEmpty(t, m.UpdatedAt.ValueString())
}

func TestPublicCloudVMFromAPINilStatus(t *testing.T) {
	vm := &pcvmv1.PublicCloudVM{
		Uuid: "uuid",
		Name: "vm",
		Size: pcvmv1.Size_SIZE_SMALL,
	}
	m := publicCloudVMFromAPI(vm)
	assert.Equal(t, "", m.StatusState.ValueString())
	assert.Equal(t, "", m.StatusOperation.ValueString())
	assert.Equal(t, "", m.StatusInfo.ValueString())
	// nil Meta -> null labels, empty updated_at.
	assert.True(t, m.Labels.IsNull())
	assert.Equal(t, "", m.UpdatedAt.ValueString())
}

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

func TestLabelsToMap(t *testing.T) {
	assert.Equal(t, map[string]string{}, labelsToMap(types.MapNull(types.StringType)))
	assert.Equal(t, map[string]string{}, labelsToMap(types.MapUnknown(types.StringType)))
	assert.Equal(t, map[string]string{"a": "1"}, labelsToMap(mapValue(t, map[string]string{"a": "1"})))
}

func TestLabelsDelta(t *testing.T) {
	// No change -> nil (nothing to do).
	assert.Nil(t, labelsDelta(types.MapNull(types.StringType), types.MapNull(types.StringType)))
	assert.Nil(t, labelsDelta(mapValue(t, map[string]string{"a": "1"}), mapValue(t, map[string]string{"a": "1"})))

	// Add only.
	add := labelsDelta(types.MapNull(types.StringType), mapValue(t, map[string]string{"a": "1"}))
	assert.NotNil(t, add)
	assert.Equal(t, map[string]string{"a": "1"}, add.Update.Labels)
	assert.Empty(t, add.Remove)

	// Remove only: state has a label the plan drops.
	rem := labelsDelta(mapValue(t, map[string]string{"a": "1", "b": "2"}), mapValue(t, map[string]string{"b": "2"}))
	assert.NotNil(t, rem)
	assert.Nil(t, rem.Update)
	assert.Equal(t, []string{"a"}, rem.Remove)

	// Add + overwrite + remove.
	both := labelsDelta(mapValue(t, map[string]string{"a": "1", "c": "3"}), mapValue(t, map[string]string{"a": "1", "b": "2", "d": "4"}))
	assert.NotNil(t, both)
	assert.Equal(t, map[string]string{"b": "2", "d": "4"}, both.Update.Labels)
	assert.Equal(t, []string{"c"}, both.Remove)

	// Plan null clears all state labels.
	clearAll := labelsDelta(mapValue(t, map[string]string{"a": "1", "b": "2"}), types.MapNull(types.StringType))
	assert.NotNil(t, clearAll)
	assert.Nil(t, clearAll.Update)
	assert.ElementsMatch(t, []string{"a", "b"}, clearAll.Remove)
}
