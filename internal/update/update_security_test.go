package update

import "testing"

func TestValidateRepo(t *testing.T) {
	valid := []string{
		"ronappleton/survive-it",
		"org.repo/name-1",
	}
	for _, repo := range valid {
		if err := validateRepo(repo); err != nil {
			t.Fatalf("expected valid repo %q, got error: %v", repo, err)
		}
	}

	invalid := []string{
		"",
		"owner",
		"owner/repo/extra",
		"owner /repo",
		"owner/repo?x=1",
		"../owner/repo",
	}
	for _, repo := range invalid {
		if err := validateRepo(repo); err == nil {
			t.Fatalf("expected invalid repo %q to fail", repo)
		}
	}
}

func TestValidateHTTPSURL(t *testing.T) {
	allowed := map[string]struct{}{
		"github.com": {},
	}

	if err := validateHTTPSURL("https://github.com/ronappleton/survive-it", allowed); err != nil {
		t.Fatalf("expected allowed URL to pass: %v", err)
	}

	if err := validateHTTPSURL("http://github.com/ronappleton/survive-it", allowed); err == nil {
		t.Fatalf("expected non-https URL to fail")
	}

	if err := validateHTTPSURL("https://example.com/ronappleton/survive-it", allowed); err == nil {
		t.Fatalf("expected non-allowlisted host URL to fail")
	}
}
