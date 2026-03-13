package sanitizer

import (
	"sync"

	"github.com/microcosm-cc/bluemonday"
)

// policy is a lazily-initialised UGC (User Generated Content) policy.
// It allows the safe subset of HTML produced by TipTap's StarterKit:
//
//   - Inline formatting: <strong>, <em>
//   - Lists: <ul>, <ol>, <li>
//   - Paragraphs: <p>
//   - Links: <a href="…"> with rel="nofollow noopener noreferrer"
var (
	once   sync.Once
	policy *bluemonday.Policy
)

func getPolicy() *bluemonday.Policy {
	once.Do(func() {
		policy = bluemonday.UGCPolicy()
		// TipTap outputs target="_blank" on links — allow it.
		policy.AllowAttrs("target").OnElements("a")
	})
	return policy
}

// SanitizeHTML sanitises untrusted HTML, keeping only the safe
// subset of tags allowed by the UGC policy. Empty or whitespace-only
// input is returned unchanged.
func SanitizeHTML(html string) string {
	if html == "" {
		return html
	}
	return getPolicy().Sanitize(html)
}
