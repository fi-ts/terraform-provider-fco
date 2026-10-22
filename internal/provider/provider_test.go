package provider

import (
	"context"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"testing"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/tfsdk"
	"github.com/hashicorp/terraform-plugin-go/tftypes"

	"github.com/fi-ts/terraform-provider-fco/internal/session"
)

// testJWT returns a minimally valid, signed JWT so the client's claim
// validation passes during Configure.
func testJWT(t *testing.T) string {
	t.Helper()

	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatalf("generate ecdsa key: %v", err)
	}

	claims := &jwt.RegisteredClaims{
		ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		Issuer:    "test",
	}

	token := jwt.NewWithClaims(jwt.SigningMethodES256, claims)
	tokenString, err := token.SignedString(key)
	if err != nil {
		t.Fatalf("sign test token: %v", err)
	}

	return tokenString
}

// newConfigureConfig builds a provider tfsdk.Config; nil fields are left null
// so environment fallbacks are exercised.
func newConfigureConfig(
	token, baseURL, projectUUID, tenant *string,
) tfsdk.Config {
	objType := tftypes.Object{
		AttributeTypes: map[string]tftypes.Type{
			"token":        tftypes.String,
			"base_url":     tftypes.String,
			"project_uuid": tftypes.String,
			"tenant":       tftypes.String,
		},
	}

	str := func(v *string) tftypes.Value {
		if v == nil {
			return tftypes.NewValue(tftypes.String, nil)
		}
		return tftypes.NewValue(tftypes.String, *v)
	}

	return tfsdk.Config{
		Raw: tftypes.NewValue(objType, map[string]tftypes.Value{
			"token":        str(token),
			"base_url":     str(baseURL),
			"project_uuid": str(projectUUID),
			"tenant":       str(tenant),
		}),
	}
}

func configure(t *testing.T, cfg tfsdk.Config) *provider.ConfigureResponse {
	t.Helper()

	p := New("test")()

	// Config.Get decodes raw values using the provider schema.
	schemaResp := &provider.SchemaResponse{}
	p.Schema(context.Background(), provider.SchemaRequest{}, schemaResp)
	cfg.Schema = schemaResp.Schema

	resp := &provider.ConfigureResponse{}
	p.Configure(context.Background(), provider.ConfigureRequest{Config: cfg}, resp)
	return resp
}

func clearFCOEnv(t *testing.T) {
	t.Helper()
	for _, k := range []string{"FCO_TOKEN", "FCO_BASE_URL", "FCO_PROJECT_UUID", "FCO_TENANT"} {
		t.Setenv(k, "")
	}
}

func TestConfigure_MissingBaseURL(t *testing.T) {
	clearFCOEnv(t)
	token := testJWT(t)
	t.Setenv("FCO_TOKEN", token)

	resp := configure(t, newConfigureConfig(nil, nil, nil, nil))
	if !resp.Diagnostics.HasError() {
		t.Fatalf("expected an error when base_url is missing, got none")
	}
}

func TestConfigure_EnvFallbackBuildsSession(t *testing.T) {
	clearFCOEnv(t)
	token := testJWT(t)
	t.Setenv("FCO_TOKEN", token)
	t.Setenv("FCO_BASE_URL", "https://api.fits.de")
	t.Setenv("FCO_PROJECT_UUID", "11111111-1111-1111-1111-111111111111")
	t.Setenv("FCO_TENANT", "env-tenant")

	resp := configure(t, newConfigureConfig(nil, nil, nil, nil))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	s, ok := resp.DataSourceData.(*session.Session)
	if !ok {
		t.Fatalf("expected *session.Session in DataSourceData, got %T", resp.DataSourceData)
	}
	if s.ProjectUUID != "11111111-1111-1111-1111-111111111111" {
		t.Errorf("expected env project_uuid, got %q", s.ProjectUUID)
	}
	if s.Tenant != "env-tenant" {
		t.Errorf("expected env tenant, got %q", s.Tenant)
	}
	if s.Client == nil {
		t.Error("expected a non-nil API client")
	}

	if resp.DataSourceData != resp.ResourceData {
		t.Error("expected DataSourceData and ResourceData to be the same *session.Session")
	}
}

func TestConfigure_ConfigWinsOverEnv(t *testing.T) {
	clearFCOEnv(t)
	envToken := testJWT(t)
	cfgToken := testJWT(t)
	t.Setenv("FCO_TOKEN", envToken)
	t.Setenv("FCO_BASE_URL", "https://env.fits.de")
	t.Setenv("FCO_PROJECT_UUID", "22222222-2222-2222-2222-222222222222")
	t.Setenv("FCO_TENANT", "env-tenant")

	cfgBaseURL := "https://cfg.fits.de"
	cfgProject := "33333333-3333-3333-3333-333333333333"
	cfgTenant := "cfg-tenant"

	resp := configure(t, newConfigureConfig(&cfgToken, &cfgBaseURL, &cfgProject, &cfgTenant))
	if resp.Diagnostics.HasError() {
		t.Fatalf("unexpected diagnostics: %v", resp.Diagnostics)
	}

	s, ok := resp.ResourceData.(*session.Session)
	if !ok {
		t.Fatalf("expected *session.Session in ResourceData, got %T", resp.ResourceData)
	}
	if s.ProjectUUID != "33333333-3333-3333-3333-333333333333" {
		t.Errorf("expected config project_uuid to win, got %q", s.ProjectUUID)
	}
	if s.Tenant != "cfg-tenant" {
		t.Errorf("expected config tenant to win, got %q", s.Tenant)
	}
	if s.Client == nil {
		t.Error("expected a non-nil API client")
	}
}
