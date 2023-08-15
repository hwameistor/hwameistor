package definitions

type helpMessage struct {
	Short string
	Long  string
}

// CmdHelpMessages [CmdGroup][CmdName]
var CmdHelpMessages = map[string]map[string]helpMessage{
	"hwameictl": {
		"hwameictl": {
			Short: "Hwameictl is the command-line tool for Hwameistor.",
			Long: "Hwameictl is a tool that can manage all Hwameistor resources and their entire lifecycle.\n" +
				"Complete documentation is available at https://hwameistor.io/",
		},
	},

	"volume": {
		"volume": {
			Short: "Manage the hwameistor's LocalVolumes.",
			Long:  "Manage the hwameistor's LocalVolumes.",
		},
		"get":     {},
		"convert": {},
		"migrate": {},
		"expand":  {},
	},
}
