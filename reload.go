package filter

func (f *Filter) Load() error {
	whitelist := NewPatternMatcher()
	blocklist := NewPatternMatcher()

	for _, list := range f.Lists {
		rc, err := list.Open()
		if err != nil {
			return err
		}
		if list.Block {
			if _, err := blocklist.ReadFrom(rc); err != nil {
				return err
			}
		} else {
			if _, err := whitelist.ReadFrom(rc); err != nil {
				return err
			}
		}
		rc.Close()
	}

	f.whitelist = whitelist
	f.blacklist = blocklist
	return nil
}
