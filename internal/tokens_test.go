package internal

import (
	"testing"
)

func TestValidateField(t *testing.T) {
	cases := []struct {
		value, field string
		wantEmpty    bool
	}{
		{"gdrive", "provider", true},
		{"myapp/production", "subdirectory", true},
		{"file-2024.tar", "name", true},
		{"", "provider", false},
		{"path/../escape", "subdirectory", false},
		{"bad chars!", "provider", false},
		{"space in name", "subdirectory", false},
	}
	for _, c := range cases {
		msg := ValidateField(c.value, c.field)
		if c.wantEmpty && msg != "" {
			t.Errorf("ValidateField(%q, %q) = %q, want empty", c.value, c.field, msg)
		}
		if !c.wantEmpty && msg == "" {
			t.Errorf("ValidateField(%q, %q) = empty, want error", c.value, c.field)
		}
	}
}

func TestParseGrant(t *testing.T) {
	t.Run("valid single permission", func(t *testing.T) {
		g, err := ParseGrant("gdrive myapp/production create")
		if err != nil {
			t.Fatal(err)
		}
		if g.Provider != "gdrive" || g.Subdirectory != "myapp/production" {
			t.Errorf("unexpected grant: %+v", g)
		}
		if len(g.Permissions) != 1 || g.Permissions[0] != "create" {
			t.Errorf("unexpected permissions: %v", g.Permissions)
		}
	})

	t.Run("valid multiple permissions", func(t *testing.T) {
		g, err := ParseGrant("s3 backups/db read,delete")
		if err != nil {
			t.Fatal(err)
		}
		if len(g.Permissions) != 2 {
			t.Errorf("expected 2 permissions, got %v", g.Permissions)
		}
	})

	t.Run("wrong field count", func(t *testing.T) {
		if _, err := ParseGrant("gdrive create"); err == nil {
			t.Error("expected error for missing subdirectory")
		}
	})

	t.Run("unknown permission", func(t *testing.T) {
		if _, err := ParseGrant("gdrive myapp write"); err == nil {
			t.Error("expected error for unknown permission 'write'")
		}
	})

	t.Run("invalid provider", func(t *testing.T) {
		if _, err := ParseGrant("bad provider! myapp read"); err == nil {
			t.Error("expected error for invalid provider")
		}
	})

	t.Run("path traversal in subdirectory", func(t *testing.T) {
		if _, err := ParseGrant("gdrive ../escape read"); err == nil {
			t.Error("expected error for path traversal in subdirectory")
		}
	})
}

func TestDeleteFromStore(t *testing.T) {
	t.Run("deletes existing token", func(t *testing.T) {
		ts := tokenStore{
			"token-a": {{Provider: "gdrive", Subdirectory: "app", Permissions: []string{"read"}}},
			"token-b": {{Provider: "s3", Subdirectory: "app", Permissions: []string{"create"}}},
		}
		if err := deleteFromStore(ts, "token-a"); err != nil {
			t.Fatal(err)
		}
		if _, ok := ts["token-a"]; ok {
			t.Error("token-a should have been deleted")
		}
		if _, ok := ts["token-b"]; !ok {
			t.Error("token-b should still exist")
		}
	})

	t.Run("errors on unknown token", func(t *testing.T) {
		ts := tokenStore{}
		if err := deleteFromStore(ts, "no-such-token"); err == nil {
			t.Error("expected error for unknown token")
		}
	})
}

func TestCheckGrants(t *testing.T) {
	ts := tokenStore{
		"token-a": {
			{Provider: "gdrive", Subdirectory: "app/prod", Permissions: []string{"create", "read"}},
		},
		"token-b": {
			{Provider: "gdrive", Subdirectory: "app/prod", Permissions: []string{"delete"}},
			{Provider: "s3", Subdirectory: "other", Permissions: []string{"read"}},
		},
	}

	cases := []struct {
		token, provider, subdirectory, permission string
		want                                       bool
	}{
		// token-a: create and read on gdrive/app/prod
		{"token-a", "gdrive", "app/prod", "create", true},
		{"token-a", "gdrive", "app/prod", "read", true},
		{"token-a", "gdrive", "app/prod", "delete", false},
		// wrong provider / subdirectory
		{"token-a", "s3", "app/prod", "read", false},
		{"token-a", "gdrive", "app/other", "read", false},
		// token-b: delete on gdrive/app/prod, read on s3/other
		{"token-b", "gdrive", "app/prod", "delete", true},
		{"token-b", "s3", "other", "read", true},
		{"token-b", "gdrive", "app/prod", "read", false},
		// unknown token
		{"no-such-token", "gdrive", "app/prod", "read", false},
		// empty token
		{"", "gdrive", "app/prod", "read", false},
	}

	for _, c := range cases {
		got := checkGrants(ts, c.token, c.provider, c.subdirectory, c.permission)
		if got != c.want {
			t.Errorf("checkGrants(%q, %q, %q, %q) = %v, want %v",
				c.token, c.provider, c.subdirectory, c.permission, got, c.want)
		}
	}
}
