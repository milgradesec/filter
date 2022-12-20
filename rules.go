package filter

type Rules struct {
	// allowlist *PatternMatcher
	// denylist  *PatternMatcher
}

func (r *Rules) Eval(name string) bool {
	return false
}
