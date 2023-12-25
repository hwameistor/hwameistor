package smart

import (
	"errors"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
	"reflect"
	"testing"
)

func TestCollector_Collect(t *testing.T) {
	totalResult, err := NewCollector().Collect()
	if err != nil {
		t.Fatal(err)
	}

	for _, deviceResult := range totalResult {
		deviceCtx := log.Fields{
			"Name":     deviceResult.Device.Name,
			"InfoName": deviceResult.Device.InfoName,
		}
		if deviceResult.Error != "" {
			log.WithError(errors.New(deviceResult.Error)).WithFields(deviceCtx).Error("Failed to collect SMART stats")
		}

		t := map[string]interface{}{}
		for _, table := range deviceResult.AtaSmartAttributes.Table {
			t["Name"] = table.Name
			t["ID"] = table.ID
			t["Worst"] = table.Worst
			t["Value"] = table.Value
			t["WhenFailed"] = table.WhenFailed
		}

		log.WithFields(log.Fields{
			"Name":       deviceResult.Device.Name,
			"InfoName":   deviceResult.Device.InfoName,
			"Attributes": t,
		}).Info("Succeed to collect SMART stats")
	}
}

func Test_ExampleParseSmartDeviceInfo(t *testing.T) {
	testCases := []struct {
		Description        string
		SmartCtlResultJson string
		ExpectSmartDevices []device
	}{
		{
			Description:        "It is an example result coming from `smartctl --scan --json` command, for smart::ScanDevice ",
			SmartCtlResultJson: "\t{\n\t\t \"json_format_version\": [\n\t\t   1,\n\t\t   0\n\t\t ],\n\t\t \"smartctl\": {\n\t\t   \"version\": [\n\t\t     7,\n\t\t     0\n\t\t   ],\n\t\t   \"svn_revision\": \"4883\",\n\t\t   \"platform_info\": \"x86_64-linux-3.10.0-1160.99.1.el7.x86_64\",\n\t\t   \"build_info\": \"(local build)\",\n\t\t   \"argv\": [\n\t\t     \"smartctl\",\n\t\t     \"--scan\",\n\t\t     \"--json\"\n\t\t   ],\n\t\t   \"exit_status\": 0\n\t\t },\n\t\t \"devices\": [\n\t\t   {\n\t\t     \"name\": \"/dev/sda\",\n\t\t     \"info_name\": \"/dev/sda\",\n\t\t     \"type\": \"scsi\",\n\t\t     \"protocol\": \"SCSI\"\n\t\t   },\n\t\t   {\n\t\t     \"name\": \"/dev/sdb\",\n\t\t     \"info_name\": \"/dev/sdb\",\n\t\t     \"type\": \"scsi\",\n\t\t     \"protocol\": \"SCSI\"\n\t\t   }\n\t\t ]\n\t\t}",
			ExpectSmartDevices: []device{
				{
					Name:     "/dev/sda",
					InfoName: "/dev/sda",
					Type:     "scsi",
					Protocol: "SCSI",
				},
				{
					Name:     "/dev/sdb",
					InfoName: "/dev/sdb",
					Type:     "scsi",
					Protocol: "SCSI",
				},
			},
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Description, func(t *testing.T) {
			var devices []device
			var jsonResult gjson.Result
			jsonResult = gjson.Parse(testCase.SmartCtlResultJson)
			for _, result := range jsonResult.Get(_SMARTDevices).Array() {
				devices = append(devices, device{
					Name:     result.Get("name").String(),
					InfoName: result.Get("info_name").String(),
					Type:     result.Get("type").String(),
					Protocol: result.Get("protocol").String(),
				})
			}
			if !reflect.DeepEqual(testCase.ExpectSmartDevices, devices) {
				t.Fatal("testing parse smartDevice failed")
			}
		})
	}
}
