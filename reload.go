package filter

import "time"

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

	f.Lock()
	f.whitelist = whitelist
	f.blacklist = blocklist
	f.Unlock()

	return nil
}

func (f *Filter) Reload() error {
	if err := f.Load(); err != nil {
		return err
	}
	tick := time.NewTicker(defaultReloadInterval)

	go func() {
		for {
			select {
			case <-tick.C:
				// Check file timestamps

			case <-f.stop:
				tick.Stop()
				return
			}
		}
	}()
	return nil
}

const defaultReloadInterval = 15 * time.Second
