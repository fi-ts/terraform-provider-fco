package project

import (
	"context"

	"connectrpc.com/connect"

	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/fi-ts/terraform-provider-fco/internal/session"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var (
	_ resource.Resource                = &ProjectResource{}
	_ resource.ResourceWithConfigure   = &ProjectResource{}
	_ resource.ResourceWithImportState = &ProjectResource{}
)

// NewProjectResource returns a new fco_project resource.
func NewProjectResource() resource.Resource {
	return &ProjectResource{}
}

// ProjectResource implements the fco_project resource. Projects are tenant
// scoped: the provider operates within a tenant and creates projects under
// that tenant, so no project_uuid is needed to manage projects.
type ProjectResource struct {
	session *session.Session
}

// Metadata implements resource.Resource.
func (*ProjectResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_project"
}

// Schema implements resource.Resource.
func (*ProjectResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a project on the fi-ts.io platform.\n\n" +
			"Projects belong to the tenant the provider is configured with, so multiple projects can be " +
			"managed within a single provider block.\n\n" +
			"The `updated_at` attribute is managed automatically by the provider and is used for " +
			"optimistic locking (CLIENT strategy, falling back to SERVER for freshly created " +
			"projects). Can be imported by UUID.",
		Attributes: projectResourceAttributes(),
	}
}

// Configure implements resource.ResourceWithConfigure.
func (r *ProjectResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// ProviderData is nil during validation (before ConfigureProvider runs) and
	// is not needed until the CRUD operations; no-op in that case.
	if req.ProviderData == nil {
		return
	}
	s, diag := shared.SessionFrom(req.ProviderData)
	resp.Diagnostics.Append(diag)
	if diag != nil {
		return
	}
	r.session = s
}

// Create implements resource.Resource.
func (r *ProjectResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan projectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	created, err := r.session.Client.Apiv1().Project().Create(ctx, &apiv1.ProjectServiceCreateRequest{
		Tenant:      r.session.Tenant,
		Name:        plan.Name.ValueString(),
		Description: plan.Description.ValueString(),
		Labels:      labelsToAPI(plan.Labels),
	})
	if err != nil {
		resp.Diagnostics.AddError("Failed to create project", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectFromAPI(created.GetProject()))...)
}

// Read implements resource.Resource.
func (r *ProjectResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state projectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	p, err := r.getProject(ctx, state.UUID.ValueString())
	if err != nil {
		if isNotFound(err) {
			tflog.Debug(ctx, "project not found, removing from state", map[string]any{
				"uuid": state.UUID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read project", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectFromAPI(p))...)
}

// Update implements resource.Resource.
func (r *ProjectResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state projectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan projectModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	reqMsg := &apiv1.ProjectServiceUpdateRequest{
		Project: state.UUID.ValueString(),
	}
	// The server always requires a non-nil UpdateMeta. The locking strategy
	// depends on whether we have a usable updated_at timestamp:
	//   - CLIENT (the default, strictest): the client supplies the timestamp
	//     it last saw; the server rejects the update if the entity changed
	//     since (stale-state protection). Requires UpdatedAt to be set.
	//   - SERVER: the server manages the lock; UpdatedAt may be omitted. This
	//     is the fallback for freshly-created/seeded projects that have no
	//     updated_at yet (only created_at).
	lockingStrategy := apiv1.OptimisticLockingStrategy_OPTIMISTIC_LOCKING_STRATEGY_SERVER
	var updatedAt *timestamppb.Timestamp
	if state.UpdatedAt.ValueString() != "" {
		ts, err := parseUpdatedAt(state.UpdatedAt.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Invalid updated_at",
				"Unable to parse the stored updated_at: "+err.Error(),
			)
			return
		}
		lockingStrategy = apiv1.OptimisticLockingStrategy_OPTIMISTIC_LOCKING_STRATEGY_CLIENT
		updatedAt = ts
	}
	reqMsg.UpdateMeta = &apiv1.UpdateMeta{
		UpdatedAt:       updatedAt,
		LockingStrategy: lockingStrategy,
	}

	// Only `description`, `avatar_url` and `labels` are mutable. `name` is
	// immutable at the API level and is guarded by the RequiresReplace plan
	// modifier, so it never reaches the Update operation.
	if !plan.Description.IsUnknown() && plan.Description.ValueString() != state.Description.ValueString() {
		description := plan.Description.ValueString()
		reqMsg.Description = &description
	}
	if !plan.AvatarURL.IsUnknown() && plan.AvatarURL.ValueString() != state.AvatarURL.ValueString() {
		avatarURL := plan.AvatarURL.ValueString()
		reqMsg.AvatarUrl = &avatarURL
	}
	if delta := labelsDelta(state.Labels, plan.Labels); delta != nil {
		reqMsg.Labels = delta
	}

	updated, err := r.session.Client.Apiv1().Project().Update(ctx, reqMsg)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeFailedPrecondition {
			resp.Diagnostics.AddError(
				"Project was modified externally",
				"Project was modified externally (stale updated_at). Re-run terraform plan to pick up the new state.",
			)
			return
		}
		resp.Diagnostics.AddError("Failed to update project", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, projectFromAPI(updated.GetProject()))...)
}

// Delete implements resource.Resource.
func (r *ProjectResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state projectModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.session.Client.Apiv1().Project().Delete(ctx, &apiv1.ProjectServiceDeleteRequest{
		Project: state.UUID.ValueString(),
	})
	// A NotFound is expected on an idempotent second delete and is not an error.
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete project", err.Error())
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *ProjectResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("uuid"), req, resp)
}

// getProject fetches a project by UUID.
func (r *ProjectResource) getProject(ctx context.Context, uuid string) (*apiv1.Project, error) {
	resp, err := r.session.Client.Apiv1().Project().Get(ctx, &apiv1.ProjectServiceGetRequest{
		Project: uuid,
	})
	if err != nil {
		return nil, err
	}
	return resp.GetProject(), nil
}

// isNotFound reports whether err is a connect NotFound code.
func isNotFound(err error) bool {
	return err != nil && connect.CodeOf(err) == connect.CodeNotFound
}
