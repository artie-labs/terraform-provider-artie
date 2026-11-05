package provider

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"terraform-provider-artie/internal/artieclient"
	"terraform-provider-artie/internal/openapi"
	"terraform-provider-artie/internal/provider/tfmodels"

	"github.com/hashicorp/terraform-plugin-framework/diag"
	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// Ensure provider defined types fully satisfy framework interfaces.
var _ resource.Resource = &KafkaAdminUserResource{}
var _ resource.ResourceWithConfigure = &KafkaAdminUserResource{}
var _ resource.ResourceWithImportState = &KafkaAdminUserResource{}

// kafkaProducerConsumerProfile is the access profile of users managed under /users rather than /admin-users.
const kafkaProducerConsumerProfile = "producer_consumer"

// How long Create waits for Strimzi to generate a new user's password, and how often it checks.
var (
	kafkaAdminUserPasswordTimeout      = 5 * time.Minute
	kafkaAdminUserPasswordPollInterval = 5 * time.Second
)

func NewKafkaAdminUserResource() resource.Resource {
	return &KafkaAdminUserResource{}
}

type KafkaAdminUserResource struct {
	client *openapi.ClientWithResponses
}

func (r *KafkaAdminUserResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_kafka_admin_user"
}

func (r *KafkaAdminUserResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Artie Managed Kafka admin user resource. This represents an admin-profile Kafka user on an `artie_kafka_cluster`, which has cluster-wide Kafka ACLs, including every replicated topic. Artie generates its SCRAM-SHA-512 password; creating one waits until the password is ready. Admin users can't be changed, so changing any argument forces a new user. Requires a company admin API key.",
		Attributes: map[string]schema.Attribute{
			"uuid":                  schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The unique identifier for this Kafka user."},
			"cluster_uuid":          schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "The UUID of the `artie_kafka_cluster` the user belongs to. Changing this forces a new user."},
			"username":              schema.StringAttribute{Required: true, PlanModifiers: []planmodifier.String{stringplanmodifier.RequiresReplace()}, MarkdownDescription: "The Kafka username. Changing this forces a new user."},
			"access_profile":        schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The user's access profile."},
			"consumer_group_prefix": schema.StringAttribute{Computed: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The prefix the user's consumer group IDs must start with."},
			"status":                schema.StringAttribute{Computed: true, MarkdownDescription: "The user's status, e.g. `pending` until its password is ready, then `active`."},
			"password":              schema.StringAttribute{Computed: true, Sensitive: true, PlanModifiers: []planmodifier.String{stringplanmodifier.UseStateForUnknown()}, MarkdownDescription: "The user's SCRAM-SHA-512 password. Clients connect to the cluster's `external_bootstrap_url` with this username and password."},
		},
	}
}

func (r *KafkaAdminUserResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	providerData, ok := req.ProviderData.(ArtieProviderData)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("Expected ArtieProviderData, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	client, err := providerData.NewOpenAPIClient()
	if err != nil {
		resp.Diagnostics.AddError("Unable to build Artie client", err.Error())
		return
	}

	r.client = client
}

func (r *KafkaAdminUserResource) GetStateData(ctx context.Context, state tfsdk.State, diagnostics *diag.Diagnostics) (tfmodels.KafkaAdminUser, bool) {
	var stateData tfmodels.KafkaAdminUser
	diagnostics.Append(state.Get(ctx, &stateData)...)
	return stateData, diagnostics.HasError()
}

// fetchPassword returns the user's password, or ok=false while the API answers 409 because it isn't ready yet.
func (r *KafkaAdminUserResource) fetchPassword(ctx context.Context, clusterUUID, userUUID string) (string, bool, error) {
	resp, err := r.client.KafkaAdminUserCredentialsWithResponse(ctx, clusterUUID, userUUID)
	if err == nil && resp.StatusCode() == http.StatusConflict {
		return "", false, nil
	}
	creds, err := artieclient.JSON200(resp, err)
	if err != nil {
		return "", false, err
	}
	return creds.Password, true, nil
}

// waitForPassword polls the credentials endpoint until Strimzi has generated the user's password.
func (r *KafkaAdminUserResource) waitForPassword(ctx context.Context, clusterUUID, userUUID string) (string, error) {
	deadline := time.Now().Add(kafkaAdminUserPasswordTimeout)
	for {
		password, ok, err := r.fetchPassword(ctx, clusterUUID, userUUID)
		if err != nil || ok {
			return password, err
		}
		if time.Now().After(deadline) {
			return "", fmt.Errorf("the user's password wasn't ready after %s; run `terraform apply` again once the user is active", kafkaAdminUserPasswordTimeout)
		}
		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-time.After(kafkaAdminUserPasswordPollInterval):
		}
	}
}

// findKafkaUser returns the user with the given UUID, or nil if the cluster has none.
func findKafkaUser(users []openapi.PayloadsKafkaUser, userUUID string) *openapi.PayloadsKafkaUser {
	for i := range users {
		if users[i].Uuid.String() == userUUID {
			return &users[i]
		}
	}
	return nil
}

// readIntoState fetches the user from List Kafka users, the only endpoint that returns a user without its password.
// The password is kept from state, or fetched if state has none (e.g. after an import).
func (r *KafkaAdminUserResource) readIntoState(ctx context.Context, clusterUUID, userUUID string, password types.String, state *tfsdk.State, diagnostics *diag.Diagnostics) {
	users, err := artieclient.JSON200(r.client.KafkaUserListWithResponse(ctx, clusterUUID))
	if err != nil {
		diagnostics.AddError("Unable to read Kafka admin user", err.Error())
		return
	}

	user := findKafkaUser(users.Items, userUUID)
	if user == nil {
		state.RemoveResource(ctx)
		return
	}
	if user.AccessProfile == kafkaProducerConsumerProfile {
		diagnostics.AddError("Not a Kafka admin user", fmt.Sprintf("Kafka user %q has the %s access profile; manage it with `artie_kafka_user` instead.", user.Username, user.AccessProfile))
		return
	}

	if !tfmodels.IsKnownAndNonEmpty(password) {
		fetched, ok, err := r.fetchPassword(ctx, clusterUUID, userUUID)
		if err != nil {
			diagnostics.AddError("Unable to read Kafka admin user credentials", err.Error())
			return
		}
		password = types.StringNull()
		if ok {
			password = types.StringValue(fetched)
		}
	}

	diagnostics.Append(state.Set(ctx, tfmodels.KafkaAdminUserFromAPIModel(*user, clusterUUID, password))...)
}

func (r *KafkaAdminUserResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var planData tfmodels.KafkaAdminUser
	resp.Diagnostics.Append(req.Plan.Get(ctx, &planData)...)
	if resp.Diagnostics.HasError() {
		return
	}

	clusterUUID := planData.ClusterUUID.ValueString()
	created, err := artieclient.JSON200(r.client.KafkaAdminUserCreateWithResponse(ctx, clusterUUID, planData.ToAPICreateRequest()))
	if err != nil {
		resp.Diagnostics.AddError("Unable to create Kafka admin user", err.Error())
		return
	}

	userUUID := created.KafkaUser.Uuid.String()
	password := ""
	if created.Password != nil {
		password = *created.Password
	} else {
		// Save the user before waiting so a timeout leaves it in state rather than orphaned.
		resp.Diagnostics.Append(resp.State.Set(ctx, tfmodels.KafkaAdminUserFromAPIModel(created.KafkaUser, clusterUUID, types.StringNull()))...)
		if password, err = r.waitForPassword(ctx, clusterUUID, userUUID); err != nil {
			resp.Diagnostics.AddError("Kafka admin user created, but its password isn't available", err.Error())
			return
		}
	}

	r.readIntoState(ctx, clusterUUID, userUUID, types.StringValue(password), &resp.State, &resp.Diagnostics)
}

func (r *KafkaAdminUserResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	stateData, hasError := r.GetStateData(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	r.readIntoState(ctx, stateData.ClusterUUID.ValueString(), stateData.UUID.ValueString(), stateData.Password, &resp.State, &resp.Diagnostics)
}

func (r *KafkaAdminUserResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	// Every argument forces a new user, so Terraform never plans an update.
	resp.Diagnostics.AddError("Kafka admin users can't be updated", "Every argument of `artie_kafka_admin_user` forces a new user. Please report this issue to the provider developers.")
}

func (r *KafkaAdminUserResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	stateData, hasError := r.GetStateData(ctx, req.State, &resp.Diagnostics)
	if hasError {
		return
	}

	// 204 means deleted; 200 means Kafka is still revoking the user, which finishes on its own. 404 means it's already gone.
	apiResp, err := r.client.KafkaAdminUserDeleteWithResponse(ctx, stateData.ClusterUUID.ValueString(), stateData.UUID.ValueString())
	if err == nil && apiResp.StatusCode() == http.StatusNotFound {
		return
	}
	if err := artieclient.CheckResponse(apiResp, err); err != nil {
		resp.Diagnostics.AddError("Unable to delete Kafka admin user", err.Error())
		return
	}
}

// parseKafkaAdminUserImportID splits an import ID of the form <cluster_uuid>/<user_uuid>.
func parseKafkaAdminUserImportID(id string) (string, string, bool) {
	clusterUUID, userUUID, ok := strings.Cut(id, "/")
	if !ok || clusterUUID == "" || userUUID == "" || strings.Contains(userUUID, "/") {
		return "", "", false
	}
	return clusterUUID, userUUID, true
}

func (r *KafkaAdminUserResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	clusterUUID, userUUID, ok := parseKafkaAdminUserImportID(req.ID)
	if !ok {
		resp.Diagnostics.AddError("Invalid import ID", fmt.Sprintf("Expected <cluster_uuid>/<user_uuid>, got: %q", req.ID))
		return
	}
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("cluster_uuid"), clusterUUID)...)
	resp.Diagnostics.Append(resp.State.SetAttribute(ctx, path.Root("uuid"), userUUID)...)
}
