package project

import (
	"reflect"
	"testing"
	"time"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestLabelsToAPI(t *testing.T) {
	t.Run("null map", func(t *testing.T) {
		if got := labelsToAPI(types.MapNull(types.StringType)); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})

	t.Run("empty map", func(t *testing.T) {
		m, _ := types.MapValue(types.StringType, map[string]attr.Value{})
		if got := labelsToAPI(m); got != nil {
			t.Errorf("expected nil, got %+v", got)
		}
	})

	t.Run("labels", func(t *testing.T) {
		m, diags := types.MapValue(types.StringType, map[string]attr.Value{
			"env": types.StringValue("dev"),
		})
		if diags.HasError() {
			t.Fatalf("unexpected diags: %v", diags)
		}
		got := labelsToAPI(m)
		if got == nil || !reflect.DeepEqual(got.Labels, map[string]string{"env": "dev"}) {
			t.Errorf("expected env=dev, got %+v", got)
		}
	})
}

func TestLabelsDelta(t *testing.T) {
	base, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"env": types.StringValue("dev"),
	})
	// No change -> nil.
	if got := labelsDelta(base, base); got != nil {
		t.Fatalf("expected nil delta, got %+v", got)
	}

	// Change env and add owner.
	changed, _ := types.MapValue(types.StringType, map[string]attr.Value{
		"env":   types.StringValue("prod"),
		"owner": types.StringValue("fco"),
	})
	got := labelsDelta(base, changed)
	if got == nil {
		t.Fatal("expected a non-nil delta")
	}
	if !reflect.DeepEqual(got.GetUpdate().GetLabels(), map[string]string{"env": "prod", "owner": "fco"}) {
		t.Errorf("unexpected update: %+v", got.GetUpdate())
	}

	// Remove a key (plan is null, state has one).
	removed := labelsDelta(base, types.MapNull(types.StringType))
	if removed == nil || removed.GetRemove() == nil {
		t.Fatalf("expected a remove delta, got %+v", removed)
	}
	if !reflect.DeepEqual(removed.GetRemove(), []string{"env"}) {
		t.Errorf("unexpected remove: %+v", removed.GetRemove())
	}
}

func TestAvatarURLFromAPI(t *testing.T) {
	if got := avatarURLFromAPI(nil); !got.IsNull() {
		t.Errorf("expected null, got %q", got.ValueString())
	}
	empty := ""
	if got := avatarURLFromAPI(&empty); !got.IsNull() {
		t.Errorf("expected null, got %q", got.ValueString())
	}
	url := "https://example.com/a.png"
	if got := avatarURLFromAPI(&url); got.ValueString() != url {
		t.Errorf("expected %q, got %q", url, got.ValueString())
	}
}

func TestParseUpdatedAt(t *testing.T) {
	if _, err := parseUpdatedAt(""); err == nil {
		t.Error("expected an error for an empty timestamp")
	}
	if _, err := parseUpdatedAt("not-a-time"); err == nil {
		t.Error("expected an error for an invalid timestamp")
	}
	in := "2024-01-01T00:00:00Z"
	got, err := parseUpdatedAt(in)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.AsTime().Format(time.RFC3339) != time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC).Format(time.RFC3339) {
		t.Errorf("unexpected round-trip: %v", got.AsTime())
	}
}

func TestProjectViewFromAPI(t *testing.T) {
	now := timestamppb.New(time.Date(2024, 2, 3, 4, 5, 6, 0, time.UTC))
	p := &apiv1.Project{
		Uuid:        "uuid-1",
		Name:        "proj",
		Description: "desc",
		Tenant:      "abc",
		Meta:        &apiv1.Meta{CreatedAt: now, UpdatedAt: now},
	}
	got := projectViewFromAPI(p)
	if got.UUID.ValueString() != "uuid-1" {
		t.Errorf("unexpected uuid: %q", got.UUID.ValueString())
	}
	if got.Name.ValueString() != "proj" {
		t.Errorf("unexpected name: %q", got.Name.ValueString())
	}
	if got.Tenant.ValueString() != "abc" {
		t.Errorf("unexpected tenant: %q", got.Tenant.ValueString())
	}
	if !got.AvatarURL.IsNull() {
		t.Errorf("expected null avatar_url, got %q", got.AvatarURL.ValueString())
	}
	if got.Labels.IsNull() == false {
		t.Errorf("expected null labels, got %v", got.Labels)
	}
}
