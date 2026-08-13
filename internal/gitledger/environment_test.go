package gitledger

import (
	"strings"
	"testing"
)

func TestControlledEnvDoesNotInheritAmbientGitVariables(t *testing.T) {
	hostile := map[string]string{
		"GIT_REFERENCE_BACKEND":              "files:///tmp/external-refs",
		"git_reference_backend":              "files:///tmp/external-refs-lower",
		"Git_Object_Directory":                "/tmp/external-objects",
		"git_alternate_object_directories":    "/tmp/external-alternates",
		"GiT_Common_Dir":                      "/tmp/external-common",
		"git_namespace":                       "hostile",
		"gIt_Config_Count":                    "1",
		"git_config_key_0":                    "core.repositoryformatversion",
		"git_config_value_0":                  "999",
		"git_replace_ref_base":                "refs/hostile/replace/",
		"git_shallow_file":                    "/tmp/hostile-shallow",
		"git_no_lazy_fetch":                   "0",
	}
	for key, value := range hostile {
		t.Setenv(key, value)
	}

	env := controlledEnv()
	allowedGit := map[string]bool{
		"GIT_CONFIG_NOSYSTEM":   true,
		"GIT_CONFIG_GLOBAL":     true,
		"GIT_TERMINAL_PROMPT":   true,
		"GIT_PAGER":             true,
		"GIT_OPTIONAL_LOCKS":    true,
		"GIT_NO_REPLACE_OBJECTS": true,
		"GIT_ATTR_NOSYSTEM":     true,
		"GIT_NO_LAZY_FETCH":     true,
	}

	for _, item := range env {
		key, _, ok := strings.Cut(item, "=")
		if !ok {
			t.Fatalf("malformed environment item %q", item)
		}
		upper := strings.ToUpper(key)
		if strings.HasPrefix(upper, "GIT_") && !allowedGit[upper] {
			t.Fatalf("unexpected inherited Git environment variable %q in controlled environment", key)
		}
		if strings.EqualFold(key, "GIT_REFERENCE_BACKEND") {
			t.Fatalf("GIT_REFERENCE_BACKEND survived controlled environment as %q", key)
		}
	}
}

func TestControlledEnvContainsRequiredFailClosedValues(t *testing.T) {
	env := controlledEnv()
	want := map[string]string{
		"GIT_CONFIG_NOSYSTEM":    "1",
		"GIT_CONFIG_GLOBAL":      "",
		"GIT_TERMINAL_PROMPT":    "0",
		"GIT_PAGER":              "cat",
		"GIT_OPTIONAL_LOCKS":     "0",
		"GIT_NO_REPLACE_OBJECTS": "1",
		"GIT_ATTR_NOSYSTEM":      "1",
		"GIT_NO_LAZY_FETCH":      "1",
		"LC_ALL":                 "C",
		"LANG":                   "C",
	}

	seen := map[string]string{}
	for _, item := range env {
		key, value, ok := strings.Cut(item, "=")
		if ok {
			seen[strings.ToUpper(key)] = value
		}
	}
	for key, expected := range want {
		got, ok := seen[key]
		if !ok {
			t.Fatalf("controlled environment missing %s", key)
		}
		if key == "GIT_CONFIG_GLOBAL" {
			if got == "" {
				t.Fatalf("GIT_CONFIG_GLOBAL must point at the platform null device")
			}
			continue
		}
		if got != expected {
			t.Fatalf("%s=%q, want %q", key, got, expected)
		}
	}
}
