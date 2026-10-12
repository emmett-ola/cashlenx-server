package verification_service

import (
	"os"
	"strings"
	"testing"
)

func TestRenderVerificationEmailMatchesReviewedFixture(t *testing.T) {
	body, err := renderVerificationEmail("123456", "30 minutes")
	if err != nil {
		t.Fatalf("renderVerificationEmail returned error: %v", err)
	}

	want, err := os.ReadFile("testdata/verification_email.html")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	if body != strings.TrimSuffix(string(want), "\n") {
		t.Fatal("rendered verification email does not match the reviewed fixture")
	}
}

func TestRenderVerificationEmailEscapesDynamicValuesAndOmitsActiveContent(t *testing.T) {
	body, err := renderVerificationEmail(`<>&"'`, `5 < 10 & "now"`)
	if err != nil {
		t.Fatalf("renderVerificationEmail returned error: %v", err)
	}

	for _, escaped := range []string{
		"&lt;&gt;&amp;&#34;&#39;",
		"5 &lt; 10 &amp; &#34;now&#34;",
	} {
		if !strings.Contains(body, escaped) {
			t.Fatalf("rendered message is missing escaped value %q", escaped)
		}
	}

	lowerBody := strings.ToLower(body)
	for _, prohibited := range []string{
		"<script",
		"<form",
		"<img",
		"<iframe",
		"http://",
		"https://",
		"tracking",
	} {
		if strings.Contains(lowerBody, prohibited) {
			t.Fatalf("rendered message contains prohibited content %q", prohibited)
		}
	}
}

func TestRenderVerificationEmailKeepsCodeSelectableTextAtNarrowWidths(t *testing.T) {
	body, err := renderVerificationEmail("123456", "30 minutes")
	if err != nil {
		t.Fatalf("renderVerificationEmail returned error: %v", err)
	}

	for _, required := range []string{
		`<meta name="viewport" content="width=device-width, initial-scale=1">`,
		`max-width:560px`,
		`aria-label="Verification code"`,
		`>123456</code>`,
	} {
		if !strings.Contains(body, required) {
			t.Fatalf("rendered message is missing %q", required)
		}
	}
}
