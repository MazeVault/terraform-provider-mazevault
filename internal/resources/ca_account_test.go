package resources

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-framework/types"
)

func TestCAAccountValidateCredentials(t *testing.T) {
	tests := []struct {
		name        string
		model       CAAccountModel
		wantErrPart string
	}{
		{
			name: "internal provider requires no credentials",
			model: CAAccountModel{
				ProviderType: types.StringValue("internal"),
			},
		},
		{
			name: "provider type rejects empty value",
			model: CAAccountModel{
				ProviderType: types.StringValue(""),
			},
			wantErrPart: "non-empty canonical lowercase",
		},
		{
			name: "provider type rejects whitespace only value",
			model: CAAccountModel{
				ProviderType: types.StringValue("   "),
			},
			wantErrPart: "non-empty canonical lowercase",
		},
		{
			name: "provider type rejects unsupported value without credentials",
			model: CAAccountModel{
				ProviderType: types.StringValue("unknownca"),
			},
			wantErrPart: "is not supported",
		},
		{
			name: "provider type rejects unsupported value with credentials",
			model: CAAccountModel{
				ProviderType: types.StringValue("unknownca"),
				APIKey:       types.StringValue("test-key"),
			},
			wantErrPart: "is not supported",
		},
		{
			name: "internal provider rejects non-canonical case",
			model: CAAccountModel{
				ProviderType: types.StringValue("Internal"),
			},
			wantErrPart: `use "internal"`,
		},
		{
			name: "internal provider rejects leading whitespace",
			model: CAAccountModel{
				ProviderType: types.StringValue(" internal"),
			},
			wantErrPart: `use "internal"`,
		},
		{
			name: "internal provider rejects trailing whitespace",
			model: CAAccountModel{
				ProviderType: types.StringValue("internal "),
			},
			wantErrPart: `use "internal"`,
		},
		{
			name: "digicert requires api key",
			model: CAAccountModel{
				ProviderType: types.StringValue("digicert"),
			},
			wantErrPart: "requires api_key",
		},
		{
			name: "digicert accepts api key",
			model: CAAccountModel{
				ProviderType: types.StringValue("digicert"),
				APIKey:       types.StringValue("test-key"),
			},
		},
		{
			name: "rapidssl legacy alias accepts api key",
			model: CAAccountModel{
				ProviderType: types.StringValue("rapidssl"),
				APIKey:       types.StringValue("test-key"),
			},
		},
		{
			name: "rapidssl legacy alias requires api key",
			model: CAAccountModel{
				ProviderType: types.StringValue("rapidssl"),
			},
			wantErrPart: "requires api_key",
		},
		{
			name: "ica placeholder rejects terraform creation",
			model: CAAccountModel{
				ProviderType: types.StringValue("ica"),
				APIKey:       types.StringValue("test-key"),
			},
			wantErrPart: "is not supported",
		},
		{
			name: "postsignum placeholder rejects terraform creation",
			model: CAAccountModel{
				ProviderType: types.StringValue("postsignum"),
				APIKey:       types.StringValue("test-key"),
			},
			wantErrPart: "is not supported",
		},
		{
			name: "psd2 qwac placeholder rejects terraform creation",
			model: CAAccountModel{
				ProviderType: types.StringValue("psd2_qwac"),
				APIKey:       types.StringValue("test-key"),
			},
			wantErrPart: "is not supported",
		},
		{
			name: "psd2 qsealc placeholder rejects terraform creation",
			model: CAAccountModel{
				ProviderType: types.StringValue("psd2_qsealc"),
				APIKey:       types.StringValue("test-key"),
			},
			wantErrPart: "is not supported",
		},
		{
			name: "adcs requires username and password in credentials map",
			model: CAAccountModel{
				ProviderType: types.StringValue("adcs"),
				APIKey:       types.StringValue("test-key"),
			},
			wantErrPart: `credentials["username"]`,
		},
		{
			name: "adcs accepts username and password in credentials map",
			model: CAAccountModel{
				ProviderType: types.StringValue("adcs"),
				Credentials:  credentialsMap(t, map[string]string{"username": "svc-enroll", "password": "s3cret"}),
			},
		},
		{
			name: "vault accepts provider specific credentials",
			model: CAAccountModel{
				ProviderType: types.StringValue("vault"),
				APIKey:       types.StringValue("test-key"),
			},
		},
		{
			name: "smallstep requires provisioner password",
			model: CAAccountModel{
				ProviderType: types.StringValue("smallstep"),
				APIKey:       types.StringValue("test-key"),
			},
			wantErrPart: "provisioner_password",
		},
		{
			name: "smallstep accepts provisioner password in credentials map",
			model: CAAccountModel{
				ProviderType: types.StringValue("smallstep"),
				Credentials:  credentialsMap(t, map[string]string{"provisioner_password": "jwk-pass"}),
			},
		},
		{
			name: "est accepts provider specific credentials",
			model: CAAccountModel{
				ProviderType: types.StringValue("est"),
				APIKey:       types.StringValue("test-key"),
			},
		},
		{
			name: "venafi accepts api key",
			model: CAAccountModel{
				ProviderType: types.StringValue("venafi"),
				APIKey:       types.StringValue("test-key"),
			},
		},
		{
			name: "digicert rejects non-canonical case",
			model: CAAccountModel{
				ProviderType: types.StringValue("DIGICERT"),
				APIKey:       types.StringValue("test-key"),
			},
			wantErrPart: `use "digicert"`,
		},
		{
			name: "digicert rejects surrounding whitespace",
			model: CAAccountModel{
				ProviderType: types.StringValue(" digicert "),
				APIKey:       types.StringValue("test-key"),
			},
			wantErrPart: `use "digicert"`,
		},
		{
			name: "lets encrypt requires email",
			model: CAAccountModel{
				ProviderType: types.StringValue("letsencrypt"),
			},
			wantErrPart: "requires email",
		},
		{
			name: "lets encrypt accepts email without directory url",
			model: CAAccountModel{
				ProviderType: types.StringValue("letsencrypt"),
				Email:        types.StringValue("pki@example.com"),
			},
		},
		{
			name: "custom acme requires directory url",
			model: CAAccountModel{
				ProviderType: types.StringValue("acme"),
				Email:        types.StringValue("pki@example.com"),
			},
			wantErrPart: "requires directory_url",
		},
		{
			name: "custom acme accepts email and directory url",
			model: CAAccountModel{
				ProviderType: types.StringValue("acme"),
				Email:        types.StringValue("pki@example.com"),
				DirectoryURL: types.StringValue("https://acme.internal.example.com/directory"),
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.model.validateCredentials(context.Background())
			if tt.wantErrPart == "" {
				if err != nil {
					t.Fatalf("validateCredentials() error = %v", err)
				}
				return
			}
			if err == nil || !strings.Contains(err.Error(), tt.wantErrPart) {
				t.Fatalf("validateCredentials() error = %v, want containing %q", err, tt.wantErrPart)
			}
		})
	}
}

func credentialsMap(t *testing.T, values map[string]string) types.Map {
	t.Helper()
	m, diags := types.MapValueFrom(context.Background(), types.StringType, values)
	if diags.HasError() {
		t.Fatalf("credentialsMap: %v", diags)
	}
	return m
}

func TestCAAccountSettingsRejectSecretLikeKeys(t *testing.T) {
	tests := []struct {
		name     string
		settings map[string]string
		wantErr  bool
	}{
		{name: "smallstep connection settings are allowed", settings: map[string]string{"server_url": "https://ca.local:9000", "provisioner_name": "admin", "root_fingerprint": "ab"}},
		{name: "provisioner password is rejected", settings: map[string]string{"provisioner_password": "x"}, wantErr: true},
		{name: "api key is rejected", settings: map[string]string{"api_key": "x"}, wantErr: true},
		{name: "client secret is rejected", settings: map[string]string{"client_secret": "x"}, wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := CAAccountModel{Settings: credentialsMap(t, tt.settings)}
			diags := model.validateSettingsHaveNoSecrets(context.Background())
			if diags.HasError() != tt.wantErr {
				t.Fatalf("validateSettingsHaveNoSecrets() hasError = %v, want %v (%v)", diags.HasError(), tt.wantErr, diags)
			}
		})
	}
}

func TestCanonicalCAProviderType(t *testing.T) {
	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "already canonical", value: "acme", want: "acme"},
		{name: "mixed case internal", value: "Internal", want: "internal"},
		{name: "commercial with whitespace", value: " DIGICERT ", want: "digicert"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := canonicalCAProviderType(tt.value); got != tt.want {
				t.Fatalf("canonicalCAProviderType(%q) = %q, want %q", tt.value, got, tt.want)
			}
		})
	}
}
