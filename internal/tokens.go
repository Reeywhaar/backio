package internal

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"regexp"
	"strings"
)

const TokensFile = "/data/tokens.json"

var safePathRE = regexp.MustCompile(`^[A-Za-z0-9._\-][A-Za-z0-9._\-/]*$`)

func ValidateField(value, field string) string {
	if value == "" {
		return field + " is required"
	}
	if strings.Contains(value, "..") {
		return field + " must not contain '..'"
	}
	if !safePathRE.MatchString(value) {
		return field + " contains invalid characters"
	}
	return ""
}

type Grant struct {
	Provider     string   `json:"provider"`
	Subdirectory string   `json:"subdirectory"`
	Permissions  []string `json:"permissions"`
}

type tokenStore map[string][]Grant

func loadTokens() (tokenStore, error) {
	data, err := os.ReadFile(TokensFile)
	if os.IsNotExist(err) {
		return tokenStore{}, nil
	}
	if err != nil {
		return nil, err
	}
	var ts tokenStore
	if err := json.Unmarshal(data, &ts); err != nil {
		return nil, err
	}
	return ts, nil
}

func saveTokens(ts tokenStore) error {
	if err := os.MkdirAll("/data", 0700); err != nil {
		return err
	}
	data, err := json.MarshalIndent(ts, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(TokensFile, data, 0600)
}

func generateToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func ParseGrant(s string) (Grant, error) {
	parts := strings.Fields(s)
	if len(parts) != 3 {
		return Grant{}, fmt.Errorf("invalid grant %q: expected \"provider subdirectory perm1,perm2\"", s)
	}
	provider, subdirectory := parts[0], parts[1]
	if msg := ValidateField(provider, "provider"); msg != "" {
		return Grant{}, fmt.Errorf("invalid provider in grant %q: %s", s, msg)
	}
	if msg := ValidateField(subdirectory, "subdirectory"); msg != "" {
		return Grant{}, fmt.Errorf("invalid subdirectory in grant %q: %s", s, msg)
	}
	perms := strings.Split(parts[2], ",")
	for _, p := range perms {
		switch p {
		case "create", "read", "delete":
		default:
			return Grant{}, fmt.Errorf("unknown permission %q in grant %q (allowed: create, read, delete)", p, s)
		}
	}
	return Grant{Provider: provider, Subdirectory: subdirectory, Permissions: perms}, nil
}

// checkGrants is the pure, testable core of token checking.
func checkGrants(ts tokenStore, token, provider, subdirectory, permission string) bool {
	grants, ok := ts[token]
	if !ok {
		return false
	}
	for _, g := range grants {
		if g.Provider != provider || g.Subdirectory != subdirectory {
			continue
		}
		for _, p := range g.Permissions {
			if p == permission {
				return true
			}
		}
	}
	return false
}

// CheckToken reports whether token has permission for the given provider+subdirectory.
// Reads from disk on each call so tokens issued via CLI take effect immediately.
func CheckToken(token, provider, subdirectory, permission string) (bool, error) {
	if token == "" {
		return false, nil
	}
	ts, err := loadTokens()
	if err != nil {
		return false, err
	}
	return checkGrants(ts, token, provider, subdirectory, permission), nil
}

func CmdIssueToken(grantStrs []string) error {
	if len(grantStrs) == 0 {
		return fmt.Errorf("usage: backio issue-token \"provider subdirectory perm1,perm2\" ...")
	}
	var grants []Grant
	for _, s := range grantStrs {
		g, err := ParseGrant(s)
		if err != nil {
			return err
		}
		grants = append(grants, g)
	}
	ts, err := loadTokens()
	if err != nil {
		return fmt.Errorf("load tokens: %w", err)
	}
	token, err := generateToken()
	if err != nil {
		return fmt.Errorf("generate token: %w", err)
	}
	ts[token] = grants
	if err := saveTokens(ts); err != nil {
		return fmt.Errorf("save tokens: %w", err)
	}
	fmt.Println(token)
	return nil
}

func deleteFromStore(ts tokenStore, token string) error {
	if _, ok := ts[token]; !ok {
		return fmt.Errorf("token not found")
	}
	delete(ts, token)
	return nil
}

func CmdDeleteToken(token string) error {
	if token == "" {
		return fmt.Errorf("usage: backio delete-token <token>")
	}
	ts, err := loadTokens()
	if err != nil {
		return fmt.Errorf("load tokens: %w", err)
	}
	if err := deleteFromStore(ts, token); err != nil {
		return err
	}
	if err := saveTokens(ts); err != nil {
		return fmt.Errorf("save tokens: %w", err)
	}
	fmt.Println("deleted")
	return nil
}

func CmdListTokens() error {
	ts, err := loadTokens()
	if err != nil {
		return fmt.Errorf("load tokens: %w", err)
	}
	if len(ts) == 0 {
		fmt.Println("(no tokens)")
		return nil
	}
	for token, grants := range ts {
		fmt.Printf("%s\n", token)
		for _, g := range grants {
			fmt.Printf("  %s %s %s\n", g.Provider, g.Subdirectory, strings.Join(g.Permissions, ","))
		}
	}
	return nil
}
