package pcvm

import (
	"context"
	"fmt"
	"strings"
	"time"

	"connectrpc.com/connect"
	pcvmv1 "github.com/fi-ts/api/go/fits/api/pcvm/v1"
	apiv1 "github.com/fi-ts/api/go/fits/api/v1"
	"github.com/fi-ts/terraform-provider-fco/internal/session"
	"github.com/fi-ts/terraform-provider-fco/internal/shared"
	"github.com/hashicorp/terraform-plugin-framework-timeouts/resource/timeouts"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

// Poll configuration for waiting on asynchronous VM operations. The timeouts
// are defaults and can be overridden via the `timeouts` block.
const (
	defaultCreateTimeout = 20 * time.Minute
	defaultUpdateTimeout = 20 * time.Minute
	pollInterval         = 5 * time.Second

	// operationActive is the terminal operation we wait for after Create/Update.
	operationActive = "active"
)

var (
	_ resource.Resource                = &PublicCloudVMResource{}
	_ resource.ResourceWithConfigure   = &PublicCloudVMResource{}
	_ resource.ResourceWithImportState = &PublicCloudVMResource{}
)

// NewPublicCloudVMResource returns a new fco_public_cloud_vm resource.
func NewPublicCloudVMResource() resource.Resource {
	return &PublicCloudVMResource{}
}

// PublicCloudVMResource implements the fco_public_cloud_vm resource.
type PublicCloudVMResource struct {
	session *session.Session
}

// Metadata implements resource.Resource.
func (*PublicCloudVMResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_public_cloud_vm"
}

// Schema implements resource.Resource.
func (*PublicCloudVMResource) Schema(ctx context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Manages a public cloud VM on the fits.cloud platform.\n\n" +
			"Only `name` and `size` can be changed after creation. `size` can only be upgraded " +
			"(the new size must be greater than or equal to the current size); downgrades are rejected.\n\n" +
			"The `updated_at` attribute is managed automatically by the provider and is used for " +
			"optimistic locking (CLIENT strategy). Can be imported as `<project_uuid>/<uuid>`.",
		Attributes: publicCloudVMResourceAttributes(),
		Blocks: map[string]schema.Block{
			"timeouts": timeouts.Block(ctx, timeouts.Opts{Create: true, Update: true}),
		},
	}
}

// Configure implements resource.ResourceWithConfigure.
func (r *PublicCloudVMResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	s, diag := shared.SessionFrom(req.ProviderData)
	resp.Diagnostics.Append(diag)
	if diag != nil {
		return
	}
	r.session = s
}

// Create implements resource.Resource.
func (r *PublicCloudVMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan publicCloudVMResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	createTimeout, diags := plan.Timeouts.Create(ctx, defaultCreateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	size, ok := sizeFromString(plan.Size.ValueString())
	if !ok {
		resp.Diagnostics.AddAttributeError(
			path.Root("size"),
			"Invalid size",
			fmt.Sprintf("size %q is not one of %v", plan.Size.ValueString(), validSizes),
		)
		return
	}

	createReq := &pcvmv1.PublicCloudVMServiceCreateRequest{
		ProjectUuid: plan.ProjectUUID.ValueString(),
		Name:        plan.Name.ValueString(),
		Labels:      shared.LabelsToAPI(plan.Labels),
		Size:        size,
	}

	created, err := r.session.Client.Apipcvmv1().PublicCloudVM().Create(ctx, createReq)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create public cloud VM", err.Error())
		return
	}

	// Persist the VM right away so Terraform tracks it even if waiting fails
	// (timeout, error state, cancellation). The resource is then tainted
	// instead of orphaned.
	resp.Diagnostics.Append(resp.State.Set(ctx, publicCloudVMResourceModel{
		publicCloudVMModel: publicCloudVMFromAPI(created.Vm),
		Timeouts:           plan.Timeouts,
	})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Then wait until the VM is active: the create response only carries the
	// transient "creating" status.
	activeVM, err := r.waitForActive(ctx, plan.ProjectUUID.ValueString(), created.Vm.GetUuid(), createTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to create public cloud VM", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, publicCloudVMResourceModel{
		publicCloudVMModel: publicCloudVMFromAPI(activeVM),
		Timeouts:           plan.Timeouts,
	})...)
}

// Read implements resource.Resource.
func (r *PublicCloudVMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state publicCloudVMResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.getVM(ctx, state.ProjectUUID.ValueString(), state.UUID.ValueString())
	if err != nil {
		if isNotFound(err) {
			tflog.Debug(ctx, "public cloud VM not found, removing from state", map[string]any{
				"uuid": state.UUID.ValueString(),
			})
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read public cloud VM", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, publicCloudVMResourceModel{
		publicCloudVMModel: publicCloudVMFromAPI(vm),
		Timeouts:           state.Timeouts,
	})...)
}

// Update implements resource.Resource.
func (r *PublicCloudVMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var state publicCloudVMResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var plan publicCloudVMResourceModel
	resp.Diagnostics.Append(req.Plan.Get(ctx, &plan)...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateTimeout, diags := plan.Timeouts.Update(ctx, defaultUpdateTimeout)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	updateMeta, err := shared.ParseUpdatedAt(state.UpdatedAt.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Missing updated_at",
			"Unable to update the public cloud VM: "+err.Error()+". Re-run `terraform plan` to refresh the state.",
		)
		return
	}

	reqMsg := &pcvmv1.PublicCloudVMServiceUpdateRequest{
		Uuid:        state.UUID.ValueString(),
		ProjectUuid: state.ProjectUUID.ValueString(),
		UpdateMeta: &apiv1.UpdateMeta{
			UpdatedAt:       updateMeta,
			LockingStrategy: apiv1.OptimisticLockingStrategy_OPTIMISTIC_LOCKING_STRATEGY_CLIENT,
		},
	}

	// Only `name`, `size` and `labels` are mutable.
	if !plan.Name.IsUnknown() && plan.Name.ValueString() != state.Name.ValueString() {
		name := plan.Name.ValueString()
		reqMsg.Name = &name
	}

	if delta := shared.LabelsDelta(state.Labels, plan.Labels); delta != nil {
		reqMsg.Labels = delta
	}

	if !plan.Size.IsUnknown() && plan.Size.ValueString() != state.Size.ValueString() {
		cur, curOK := sizeFromString(state.Size.ValueString())
		next, nextOK := sizeFromString(plan.Size.ValueString())
		if !nextOK {
			resp.Diagnostics.AddAttributeError(
				path.Root("size"),
				"Invalid size",
				fmt.Sprintf("size %q is not one of %v", plan.Size.ValueString(), validSizes),
			)
			return
		}
		if curOK && isSizeDowngrade(cur, next) {
			resp.Diagnostics.AddAttributeError(
				path.Root("size"),
				"Size downgrades are not allowed",
				fmt.Sprintf("Cannot change size from %q to %q. Public cloud VMs can only be upgraded, not downgraded.", state.Size.ValueString(), plan.Size.ValueString()),
			)
			return
		}
		reqMsg.Size = &next
	}

	updated, err := r.session.Client.Apipcvmv1().PublicCloudVM().Update(ctx, reqMsg)
	if err != nil {
		if connect.CodeOf(err) == connect.CodeFailedPrecondition {
			resp.Diagnostics.AddError(
				"Public cloud VM was modified externally",
				"VM was modified externally (stale updated_at). Re-run terraform plan to pick up the new state.",
			)
			return
		}
		resp.Diagnostics.AddError("Failed to update public cloud VM", err.Error())
		return
	}

	// Persist the updated VM right away (fresh updated_at for optimistic
	// locking) so state stays correct even if waiting fails.
	resp.Diagnostics.Append(resp.State.Set(ctx, publicCloudVMResourceModel{
		publicCloudVMModel: publicCloudVMFromAPI(updated.Vm),
		Timeouts:           plan.Timeouts,
	})...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Then wait until the VM is active: the update response only carries the
	// transient "updating" status.
	activeVM, err := r.waitForActive(ctx, state.ProjectUUID.ValueString(), updated.Vm.GetUuid(), updateTimeout)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update public cloud VM", err.Error())
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, publicCloudVMResourceModel{
		publicCloudVMModel: publicCloudVMFromAPI(activeVM),
		Timeouts:           plan.Timeouts,
	})...)
}

// Delete implements resource.Resource.
func (r *PublicCloudVMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state publicCloudVMResourceModel
	resp.Diagnostics.Append(req.State.Get(ctx, &state)...)
	if resp.Diagnostics.HasError() {
		return
	}

	_, err := r.session.Client.Apipcvmv1().PublicCloudVM().Delete(ctx, &pcvmv1.PublicCloudVMServiceDeleteRequest{
		Uuid:        state.UUID.ValueString(),
		ProjectUuid: state.ProjectUUID.ValueString(),
	})
	// A NotFound is expected on an idempotent second delete and is not an error.
	if err != nil && !isNotFound(err) {
		resp.Diagnostics.AddError("Failed to delete public cloud VM", err.Error())
	}
}

// ImportState implements resource.ResourceWithImportState.
func (r *PublicCloudVMResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	projectUUID, uuid, ok := strings.Cut(req.ID, "/")
	if !ok || projectUUID == "" || uuid == "" {
		resp.Diagnostics.AddError(
			"Invalid import ID",
			fmt.Sprintf("Expected import ID in the format <project_uuid>/<uuid>, got %q.", req.ID),
		)
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("project_uuid"), projectUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uuid"), uuid)...)
}

// getVM fetches a VM by UUID in the given project.
func (r *PublicCloudVMResource) getVM(ctx context.Context, projectUUID, uuid string) (*pcvmv1.PublicCloudVM, error) {
	resp, err := r.session.Client.Apipcvmv1().PublicCloudVM().Get(ctx, &pcvmv1.PublicCloudVMServiceGetRequest{
		Uuid:        uuid,
		ProjectUuid: projectUUID,
	})
	if err != nil {
		return nil, err
	}
	return resp.Vm, nil
}

// waitForActive polls until the VM's status.operation reaches operationActive,
// an error is observed, or the deadline is reached. It returns the active VM
// so callers can persist the terminal status and includes the last
// status.info in any returned error.
func (r *PublicCloudVMResource) waitForActive(ctx context.Context, projectUUID, uuid string, timeout time.Duration) (*pcvmv1.PublicCloudVM, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	lastInfo := ""
	for {
		vm, err := r.getVM(ctx, projectUUID, uuid)
		if err != nil {
			return nil, fmt.Errorf("polling public cloud VM %s: %w", uuid, err)
		}
		if vm.GetStatus() != nil {
			if info := vm.Status.Info; info != "" {
				lastInfo = info
			}
			op := operationToString(vm.Status.Operation)
			if op == operationActive {
				tflog.Debug(ctx, "public cloud VM reached active", map[string]any{"uuid": uuid})
				return vm, nil
			}
			if op == "error" {
				return nil, fmt.Errorf("public cloud VM %s entered error state: %s", uuid, lastInfo)
			}
			tflog.Debug(ctx, "public cloud VM still in operation", map[string]any{
				"uuid":      uuid,
				"operation": op,
				"info":      lastInfo,
			})
		}

		select {
		case <-ctx.Done():
			if lastInfo != "" {
				return nil, fmt.Errorf("timed out waiting for public cloud VM %s to become active (last status: %s)", uuid, lastInfo)
			}
			return nil, fmt.Errorf("timed out waiting for public cloud VM %s to become active", uuid)
		case <-time.After(pollInterval):
		}
	}
}

// isNotFound reports whether err is a connect NotFound code.
func isNotFound(err error) bool {
	return err != nil && connect.CodeOf(err) == connect.CodeNotFound
}
