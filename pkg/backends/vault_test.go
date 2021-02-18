package backends_test

import (
	"reflect"
	"testing"

	"github.com/IBM/argocd-vault-plugin/pkg/auth/vault"
	"github.com/IBM/argocd-vault-plugin/pkg/backends"
	"github.com/IBM/argocd-vault-plugin/pkg/helpers"
)

func TestVaultLogin(t *testing.T) {
	cluster, roleID, secretID := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	backend := &backends.Vault{
		VaultClient: cluster.Cores[0].Client,
	}

	t.Run("will authenticate with approle", func(t *testing.T) {
		backend.AuthType = vault.NewAppRoleAuth(roleID, secretID)

		err := backend.Login()
		if err != nil {
			t.Fatalf("expected no errors but got: %s", err)
		}
	})
}

func TestVaultGetSecrets(t *testing.T) {
	cluster, roleID, secretID := helpers.CreateTestAppRoleVault(t)
	defer cluster.Cleanup()

	auth := vault.NewAppRoleAuth(roleID, secretID)
	backend := backends.NewVaultBackend(auth, cluster.Cores[0].Client, "")

	t.Run("will get data from vault with kv1", func(t *testing.T) {
		data, err := backend.GetSecrets("secret/foo", "1")
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"secret": "bar",
		}

		if !reflect.DeepEqual(data, expected) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("will get data from vault with kv2", func(t *testing.T) {
		data, err := backend.GetSecrets("kv/data/test", "2")
		if err != nil {
			t.Fatalf("expected 0 errors but got: %s", err)
		}

		expected := map[string]interface{}{
			"hello": "world",
		}

		if !reflect.DeepEqual(data, expected) {
			t.Errorf("expected: %s, got: %s.", expected, data)
		}
	})

	t.Run("will throw an error if cant find secrets", func(t *testing.T) {
		_, err := backend.GetSecrets("kv/data/no_path", "2")
		if err == nil {
			t.Fatalf("expected an error but did not get an error")
		}

		expected := "Could not find secrets at path kv/data/no_path"

		if !reflect.DeepEqual(err.Error(), expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

	t.Run("will throw an error if cant find secrets", func(t *testing.T) {
		_, err := backend.GetSecrets("kv/data/bad_test", "2")
		if err == nil {
			t.Fatalf("expected an error but did not get an error")
		}

		expected := "Could not get data from Vault, check that kv-v2 is the correct engine"

		if !reflect.DeepEqual(err.Error(), expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

	t.Run("will throw an error if unsupported kv version", func(t *testing.T) {
		_, err := backend.GetSecrets("kv/data/test", "3")
		if err == nil {
			t.Fatalf("expected an error but did not get an error")
		}

		expected := "Unsupported kvVersion specified"

		if !reflect.DeepEqual(err.Error(), expected) {
			t.Errorf("expected: %s, got: %s.", expected, err.Error())
		}
	})

}
