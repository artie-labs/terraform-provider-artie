package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"regexp"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"
	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/knownvalue"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/hashicorp/terraform-plugin-testing/statecheck"
	"github.com/hashicorp/terraform-plugin-testing/tfjsonpath"
	"github.com/stretchr/testify/assert"

	"terraform-provider-artie/internal/openapi"
)

// previousProviderVersion is the latest published release, used to write state the way existing users have it.
const previousProviderVersion = "2.3.50"

var localProvider = map[string]func() (tfprotov6.ProviderServer, error){
	"artie": providerserver.NewProtocol6WithError(New("test")()),
}

var partitionByPath = tfjsonpath.New("tables").AtMapKey("public.events").AtMapKey("merge_predicates").AtSliceIndex(0).AtMapKey("partition_by")

// fakeArtieAPI serves the pipeline endpoints the pipeline resource calls. Like the API, an update replaces the stored
// pipeline with the request, except that a time merge predicate sent without partitionBy keeps the stored one.
type fakeArtieAPI struct {
	*httptest.Server
	mu       sync.Mutex
	pipeline map[string]any
	// sentPartitionBy is the first merge predicate's partitionBy in the last create/update request, nil if left out.
	sentPartitionBy any
}

func newFakeArtieAPI(t *testing.T) *fakeArtieAPI {
	api := &fakeArtieAPI{}
	mux := http.NewServeMux()
	ok := func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{}`))
	}
	mux.HandleFunc("POST /pipelines/validate-unsaved-source", ok)
	mux.HandleFunc("POST /pipelines/validate-unsaved-destination", ok)
	mux.HandleFunc("POST /pipelines/{uuid}/start", ok)
	mux.HandleFunc("DELETE /pipelines/{uuid}", ok)
	mux.HandleFunc("POST /pipelines", api.save)
	mux.HandleFunc("POST /pipelines/{uuid}", api.save)
	mux.HandleFunc("GET /pipelines/{uuid}", func(w http.ResponseWriter, _ *http.Request) {
		api.mu.Lock()
		defer api.mu.Unlock()
		api.write(t, w)
	})
	api.Server = httptest.NewServer(mux)
	t.Cleanup(api.Close)
	return api
}

func (a *fakeArtieAPI) save(w http.ResponseWriter, r *http.Request) {
	var body struct {
		Pipeline map[string]any `json:"pipeline"`
	}
	if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	a.mu.Lock()
	defer a.mu.Unlock()
	pipeline := body.Pipeline
	pipeline["uuid"] = uuid.NewString()
	if a.pipeline != nil {
		pipeline["uuid"] = a.pipeline["uuid"]
	}
	for key, value := range map[string]any{
		"companyUUID": uuid.NewString(), "environmentUUID": uuid.NewString(), "createdAt": time.Now(), "lastUpdatedAt": time.Now(),
		"sourceType": "postgresql", "hasUndeployedChanges": false, "isDeploying": false, "status": "running",
	} {
		pipeline[key] = value
	}
	settings, _ := pipeline["advancedSettings"].(map[string]any)
	if settings == nil {
		settings = map[string]any{}
		pipeline["advancedSettings"] = settings
	}
	// The API always returns the flush rules, filling in its defaults.
	for key, value := range map[string]any{"flushIntervalSeconds": 60, "bufferRows": 5000, "flushSizeKb": 25000} {
		if settings[key] == nil {
			settings[key] = value
		}
	}
	tables, _ := pipeline["tables"].([]any)
	for _, table := range tables {
		if table, _ := table.(map[string]any); table["uuid"] == nil {
			table["uuid"] = uuid.NewString()
		}
	}
	sent := firstMergePredicate(pipeline)
	a.sentPartitionBy = sent["partitionBy"]
	if a.pipeline != nil && sent["partitionBy"] == nil && sent["partitionType"] != "integer" {
		if stored := firstMergePredicate(a.pipeline); stored["partitionField"] == sent["partitionField"] {
			sent["partitionBy"] = stored["partitionBy"]
		}
	}
	a.pipeline = pipeline
	a.write(nil, w)
}

func (a *fakeArtieAPI) write(t *testing.T, w http.ResponseWriter) {
	if a.pipeline == nil {
		http.NotFound(w, nil)
		return
	}
	out, err := json.Marshal(a.pipeline)
	if err == nil {
		// Fail loudly if the fake drifts from the API's response shape.
		err = json.Unmarshal(out, &openapi.PayloadsFullPipeline{})
	}
	if err != nil {
		if t != nil {
			t.Error(err)
		}
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_, _ = w.Write(out)
}

func firstMergePredicate(pipeline map[string]any) map[string]any {
	tables, _ := pipeline["tables"].([]any)
	table, _ := tables[0].(map[string]any)
	settings, _ := table["advancedSettings"].(map[string]any)
	preds, _ := settings["mergePredicates"].([]any)
	pred, _ := preds[0].(map[string]any)
	return pred
}

// setPartitionBy changes the stored partitionBy the way the dashboard does, outside Terraform.
func (a *fakeArtieAPI) setPartitionBy(value string) {
	a.mu.Lock()
	defer a.mu.Unlock()
	firstMergePredicate(a.pipeline)["partitionBy"] = value
}

func (a *fakeArtieAPI) storedPartitionBy() any {
	a.mu.Lock()
	defer a.mu.Unlock()
	return firstMergePredicate(a.pipeline)["partitionBy"]
}

func (a *fakeArtieAPI) lastSentPartitionBy() any {
	a.mu.Lock()
	defer a.mu.Unlock()
	return a.sentPartitionBy
}

// pipelineConfig is a BigQuery pipeline with one merge predicate; extra is appended inside the predicate.
func pipelineConfig(endpoint, partitionType, extra string) string {
	return fmt.Sprintf(`
provider "artie" {
  endpoint = %q
  api_key  = "arsk_test"
}

resource "artie_pipeline" "test" {
  name                       = "Postgres to BigQuery"
  source_reader_uuid         = "4e6f1c1a-0b7e-4f4e-9a52-1f5b8a1c2d3e"
  destination_connector_uuid = "7a1d2c3b-4e5f-4a6b-8c7d-9e0f1a2b3c4d"
  destination_config = {
    database = "analytics"
  }
  tables = {
    "public.events" = {
      name   = "events"
      schema = "public"
      merge_predicates = [{
        partition_field = "created_at"
        partition_type  = %q
        %s
      }]
    }
  }
}
`, endpoint, partitionType, extra)
}

func TestPipelinePartitionBy_UpgradeFromPreviousRelease(t *testing.T) {
	{
		// State written by the previous release has no partition_by; the new provider plans no changes.
		api := newFakeArtieAPI(t)
		config := pipelineConfig(api.URL, "time", "")
		resource.UnitTest(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					ExternalProviders: map[string]resource.ExternalProvider{"artie": {Source: "artie-labs/artie", VersionConstraint: previousProviderVersion}},
					Config:            config,
				},
				{
					ProtoV6ProviderFactories: localProvider,
					Config:                   config,
					ConfigPlanChecks:         resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
				},
			},
		})
		assert.Nil(t, api.storedPartitionBy())
	}
	{
		// A granularity set in the dashboard after the last apply with the previous release survives refresh, plan
		// and apply with the new provider, even though the config doesn't mention it.
		api := newFakeArtieAPI(t)
		config := pipelineConfig(api.URL, "time", "")
		resource.UnitTest(t, resource.TestCase{
			Steps: []resource.TestStep{
				{
					ExternalProviders: map[string]resource.ExternalProvider{"artie": {Source: "artie-labs/artie", VersionConstraint: previousProviderVersion}},
					Config:            config,
				},
				{
					PreConfig:                func() { api.setPartitionBy("monthly") },
					ProtoV6ProviderFactories: localProvider,
					Config:                   config,
					ConfigPlanChecks:         resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectEmptyPlan()}},
					ConfigStateChecks:        []statecheck.StateCheck{statecheck.ExpectKnownValue("artie_pipeline.test", partitionByPath, knownvalue.StringExact("monthly"))},
				},
				{
					// Force an update through a different field: partition_by is still sent back unchanged.
					ProtoV6ProviderFactories: localProvider,
					Config:                   regexp.MustCompile(`Postgres to BigQuery`).ReplaceAllString(config, "Postgres to BigQuery (renamed)"),
					ConfigPlanChecks:         resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectKnownValue("artie_pipeline.test", partitionByPath, knownvalue.StringExact("monthly"))}},
					ConfigStateChecks:        []statecheck.StateCheck{statecheck.ExpectKnownValue("artie_pipeline.test", partitionByPath, knownvalue.StringExact("monthly"))},
				},
			},
		})
		assert.Equal(t, "monthly", api.lastSentPartitionBy())
		assert.Equal(t, "monthly", api.storedPartitionBy())
	}
}

func TestPipelinePartitionBy_OmittedOrNull(t *testing.T) {
	for _, extra := range []string{"", "partition_by = null"} {
		{
			// Integer predicate: plans null and is left out of the request.
			api := newFakeArtieAPI(t)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: localProvider,
				Steps: []resource.TestStep{{
					Config:            pipelineConfig(api.URL, "integer", extra),
					ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectKnownValue("artie_pipeline.test", partitionByPath, knownvalue.Null())}},
					ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue("artie_pipeline.test", partitionByPath, knownvalue.Null())},
				}},
			})
			assert.Nil(t, api.lastSentPartitionBy(), extra)
		}
		{
			// Time predicate with nothing in state: unknown, left out of the request, and read back from the API.
			api := newFakeArtieAPI(t)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: localProvider,
				Steps: []resource.TestStep{{
					Config:            pipelineConfig(api.URL, "time", extra),
					ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectUnknownValue("artie_pipeline.test", partitionByPath)}},
					ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue("artie_pipeline.test", partitionByPath, knownvalue.Null())},
				}},
			})
			assert.Nil(t, api.lastSentPartitionBy(), extra)
		}
		{
			// Time predicate with a dashboard-set value in state: kept and sent back.
			api := newFakeArtieAPI(t)
			config := pipelineConfig(api.URL, "time", extra)
			resource.UnitTest(t, resource.TestCase{
				ProtoV6ProviderFactories: localProvider,
				Steps: []resource.TestStep{
					{Config: config},
					{
						PreConfig: func() { api.setPartitionBy("hourly") },
						Config:    regexp.MustCompile(`Postgres to BigQuery`).ReplaceAllString(config, "Postgres to BigQuery (renamed)"),
						ConfigPlanChecks: resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{
							plancheck.ExpectKnownValue("artie_pipeline.test", partitionByPath, knownvalue.StringExact("hourly")),
						}},
					},
				},
			})
			assert.Equal(t, "hourly", api.lastSentPartitionBy(), extra)
		}
	}
}

func TestPipelinePartitionBy_EmptyStringIsRejected(t *testing.T) {
	api := newFakeArtieAPI(t)
	resource.UnitTest(t, resource.TestCase{
		ProtoV6ProviderFactories: localProvider,
		Steps: []resource.TestStep{{
			Config:      pipelineConfig(api.URL, "time", `partition_by = ""`),
			PlanOnly:    true,
			ExpectError: regexp.MustCompile(`value must\s+be one of`),
		}},
	})
}

func TestPipelinePartitionBy_ApplyWithoutRefresh(t *testing.T) {
	// State has no partition_by (as after upgrading, see above) while the API has "monthly", and the plan skips the
	// refresh that would have read it, so partition_by is planned unknown and left out of the request. The API keeps
	// the stored value for a field left out, so the table stays monthly. State comes from this provider rather than
	// the published release because that release predates other pipeline attributes, which a -refresh=false plan
	// can't fill in either.
	api := newFakeArtieAPI(t)
	config := pipelineConfig(api.URL, "time", "")
	resource.UnitTest(t, resource.TestCase{
		AdditionalCLIOptions:     &resource.AdditionalCLIOptions{Plan: resource.PlanOptions{NoRefresh: true}},
		ProtoV6ProviderFactories: localProvider,
		Steps: []resource.TestStep{
			{Config: config},
			{
				PreConfig:         func() { api.setPartitionBy("monthly") },
				Config:            regexp.MustCompile(`Postgres to BigQuery`).ReplaceAllString(config, "Postgres to BigQuery (renamed)"),
				ConfigPlanChecks:  resource.ConfigPlanChecks{PreApply: []plancheck.PlanCheck{plancheck.ExpectUnknownValue("artie_pipeline.test", partitionByPath)}},
				ConfigStateChecks: []statecheck.StateCheck{statecheck.ExpectKnownValue("artie_pipeline.test", partitionByPath, knownvalue.StringExact("monthly"))},
			},
		},
	})
	assert.Nil(t, api.lastSentPartitionBy())
	assert.Equal(t, "monthly", api.storedPartitionBy())
}
