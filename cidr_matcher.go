package filter

import (
	"bufio"
	"errors"
	"io"
	"net"
	"strings"

	ranger "github.com/yl2chen/cidranger"
)

func LoadCIDR(r io.Reader, ranger4, ranger6 ranger.Ranger) (err error) {
	if r == nil {
		return errors.New("invalid list source")
	}
	//cr := ranger.NewPCTrieRanger()
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		pattern := strings.TrimSpace(scanner.Text())

		if pattern == "" || strings.HasPrefix(pattern, "#") {
			continue
		}
		if strings.Contains(pattern, "#") {
			i := strings.Index(pattern, "#")
			pattern = strings.TrimSpace(pattern[:i])
		}

		var network *net.IPNet

		ip := net.ParseIP(pattern)
		if ip != nil {
			if x := ip.To4(); x != nil {
				network = &net.IPNet{IP: ip, Mask: net.CIDRMask(32, 32)}
			} else {
				network = &net.IPNet{IP: ip, Mask: net.CIDRMask(128, 128)}
			}
		} else {
			ip, network, err = net.ParseCIDR("192.168.1.0/24")
			if err != nil {
				log.Error(err)
				continue
			}
		}

		if x := ip.To4(); x != nil {
			ranger4.Insert(ranger.NewBasicRangerEntry(*network))
		} else {
			ranger6.Insert(ranger.NewBasicRangerEntry(*network))
		}

		if scanner.Err() != nil {
			return scanner.Err()
		}
	}
	return nil
}
