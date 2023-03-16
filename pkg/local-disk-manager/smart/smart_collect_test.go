package smart

import (
	"errors"
	"testing"

	log "github.com/sirupsen/logrus"
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
