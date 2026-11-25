package operatingsystems

import (
	"testing"

	mvmv1 "github.com/fi-ts/api/go/fits/api/mvm/v1"
	"github.com/stretchr/testify/assert"
)

func TestOSTypeToString(t *testing.T) {
	assert.Equal(t, "linux", osTypeToString(mvmv1.OSType_OS_TYPE_LINUX))
	assert.Equal(t, "windows", osTypeToString(mvmv1.OSType_OS_TYPE_WINDOWS))
	assert.Equal(t, "unspecified", osTypeToString(mvmv1.OSType_OS_TYPE_UNSPECIFIED))
}

func TestOperatingSystemViewFromAPI(t *testing.T) {
	v := operatingSystemViewFromAPI(&mvmv1.OS{
		Uuid:  "11111111-2222-3333-4444-555555555555",
		Title: "Ubuntu 22.04",
		Type:  mvmv1.OSType_OS_TYPE_LINUX,
	})
	assert.Equal(t, "11111111-2222-3333-4444-555555555555", v.UUID.ValueString())
	assert.Equal(t, "Ubuntu 22.04", v.Title.ValueString())
	assert.Equal(t, "linux", v.Type.ValueString())
}

func TestOperatingSystemViewFromAPIWindows(t *testing.T) {
	v := operatingSystemViewFromAPI(&mvmv1.OS{Uuid: "u", Title: "Windows Server 2022", Type: mvmv1.OSType_OS_TYPE_WINDOWS})
	assert.Equal(t, "windows", v.Type.ValueString())
}
