package project

import (
	"testing"
	"time"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"google.golang.org/protobuf/types/known/timestamppb"
)

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
	if got.Labels.IsNull() == false {
		t.Errorf("expected null labels, got %v", got.Labels)
	}
}
