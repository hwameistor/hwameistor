package udev

import "github.com/pilebones/go-udev/netlink"

// GenRuleForBlock
func GenRuleForBlock() netlink.Matcher {
	return &netlink.RuleDefinitions{
		Rules: []netlink.RuleDefinition{
			{
				Env: map[string]string{
					"SUBSYSTEM": "block",
				},
			},
		},
	}
}
