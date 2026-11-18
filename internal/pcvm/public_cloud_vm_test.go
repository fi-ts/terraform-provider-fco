package pcvm

import (
	"context"
	"testing"
	"time"

	pcvmv1 "github.com/fi-ts/api/go/fits/api/pcvm/v1"
	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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
	assert.Equal(t, "healthy", statusAttr(t, m.Status, "state"))
	assert.Equal(t, "active", statusAttr(t, m.Status, "operation"))
	assert.Equal(t, "all good", statusAttr(t, m.Status, "info"))
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
	assert.True(t, m.Status.IsNull())
	// nil Meta -> null labels, empty updated_at.
	assert.True(t, m.Labels.IsNull())
	assert.Equal(t, "", m.UpdatedAt.ValueString())
}

// statusAttr returns the string value of a nested status attribute.
func statusAttr(t *testing.T, status types.Object, key string) string {
	t.Helper()
	v, ok := status.Attributes()[key].(types.String)
	require.True(t, ok, "status.%s is not a string", key)
	return v.ValueString()
}
