package smart

import (
	"fmt"
	"path"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"

	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/disk/manager"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
)

const (
	_SMARTCtl = "smartctl"

	// json path in smartctl result
	_SMARTExitStatus   = "smartctl.exit_status"
	_SMARTMessages     = "smartctl.messages"
	_SMARTDevices      = "devices"
	_SMARTStatusPassed = "smart_status.passed"
)

type Device struct {
	Device string
	Serial string
	Model  string
}

type controller struct {
	*manager.DiskIdentify
	Options []string
}

// NewSMARTController used to get S.M.A.R.T info
func NewSMARTController(device *manager.DiskIdentify, options ...string) *controller {
	return &controller{
		DiskIdentify: device,
		Options:      options,
	}
}

// NewSMARTParser used to get S.M.A.R.T info
func NewSMARTParser(device *manager.DiskIdentify, options ...string) *manager.SmartInfoParser {
	return &manager.SmartInfoParser{
		ISmart: NewSMARTController(device, options...),
	}
}

// ParseSmartInfo gets S.M.A.R.T info, including whether the disk supports SMART technology
// and whether it has passed the health check
func (c *controller) ParseSmartInfo() manager.SmartInfo {
	var (
		healthOK bool
		err      error
		result   manager.SmartInfo
	)

	if healthOK, err = c.GetHealthStatus(); err != nil {
		log.WithError(err).Error("Failed to get disk status")
		result.OverallHealthPassed = false
	} else {
		result.OverallHealthPassed = healthOK
	}

	log.WithFields(log.Fields{"disk": c.DevName, "status": healthOK}).Info("Succeed to check disk status")
	return result
}

// SupportSmart judge if this device supports SMART or not
func (c *controller) SupportSmart() (bool, error) {
	var (
		jsonStatus gjson.Result
		err        error
	)

	if jsonStatus, err = getSMARTCtlResult(c.FormatDevPath(), append(c.Options, "-i")...); err != nil {
		return false, err
	}

	return jsonStatus.Get("").Bool(), err
}

// GetHealthStatus Show device SMART health status
// true: passed false: not passed
func (c *controller) GetHealthStatus() (bool, error) {
	var (
		jsonStatus gjson.Result
		err        error
	)

	if jsonStatus, err = getSMARTCtlResult(c.FormatDevPath(), append(c.Options, "-H")...); err != nil {
		return false, err
	}

	return jsonStatus.Get(_SMARTStatusPassed).Bool(), err
}

// GetSmartStatus represent overall health status
func (c *controller) GetSmartStatus() (bool, error) {
	var (
		jsonStatus gjson.Result
		err        error
	)

	if jsonStatus, err = getSMARTCtlResult(c.FormatDevPath(), append(c.Options, "/c0", "show")...); err != nil {
		return false, err
	}

	return jsonStatus.Get(_SMARTStatusPassed).Bool(), err
}

// FormatDevPath sda => /dev/sda
func (c *controller) FormatDevPath() string {
	if strings.HasPrefix(c.DevName, "/dev") {
		return c.DevName
	}
	return path.Join("/dev", c.Name)
}

// GetAllStats get all stats collect by smartctl
func (c *controller) GetAllStats() (gjson.Result, error) {
	return getSMARTCtlResult(c.FormatDevPath(), append(c.Options, "--all")...)
}

// getSMARTCtlResult get smartctl output
func getSMARTCtlResult(device string, options ...string) (gjson.Result, error) {
	var (
		result gjson.Result
		args   []string
	)

	if device != "" {
		args = append(options, device, "--json")
	} else {
		args = append(options, "--json")
	}

	// dismiss error for broken disk will cause non-zero exit_status too
	out, err := utils.BashWithArgs(_SMARTCtl, args...)
	if !gjson.Valid(out) {
		log.Errorf("invalid json format: %v", out)
		return result, fmt.Errorf("invalid json format: %v", err)
	}

	result = gjson.Parse(out)
	if err = resultCodeIsOk(result.Get(_SMARTExitStatus).Int()); err != nil {
		// dismiss exit_status
		log.Warningf("Abnormal information: %v", err)
	}

	if err = jsonIsOk(result); err != nil {
		log.WithError(err).Error("got error message")
		return result, err
	}

	return result, nil
}

// resultCodeIsOk parses smartctl return code and wraps it into an error or nil.
// The return values of smartctl are defined by a bitmask. If all is well with the disk,
// the return value (exist status) of smartctl is 0 (all bits turned off). If a problem occurs,
// or an error, potenitial error, or fault is detected, then a non-zero status is returned.
// More info: https://linux.die.net/man/8/smartctl
func resultCodeIsOk(SMARTCtlResult int64) error {
	var (
		err error
	)
	if SMARTCtlResult > 0 {
		b := SMARTCtlResult
		if (b & 1) != 0 {
			err = fmt.Errorf("command line did not parse")
		}
		if (b & (1 << 1)) != 0 {
			err = fmt.Errorf("device open failed, device did not return an IDENTIFY DEVICE structure, " +
				"or device is in a low-power mode")
		}
		if (b & (1 << 2)) != 0 {
			err = fmt.Errorf("some SMART or other ATA command to the disk failed, " +
				"or there was a checksum error in a SMART data structure")
		}
		if (b & (1 << 3)) != 0 {
			err = fmt.Errorf("SMART status check returned 'DISK FAILING'")
		}
		if (b & (1 << 4)) != 0 {
			err = fmt.Errorf("we found prefail Attributes <= threshold")
		}
		if (b & (1 << 5)) != 0 {
			err = fmt.Errorf("SMART status check returned 'DISK OK' but we found that " +
				"some (usage or prefail) Attributes have been <= threshold at some time in the past")
		}
		if (b & (1 << 6)) != 0 {
			err = fmt.Errorf("the device error log contains records of errors")
		}
		if (b & (1 << 7)) != 0 {
			err = fmt.Errorf("the device self-test log contains records of errors. " +
				"[ATA only] Failed self-tests outdated by a newer successful extended self-test are ignored")
		}
	}
	return err
}

// Check json
func jsonIsOk(json gjson.Result) error {
	messages := json.Get(_SMARTMessages)
	if messages.Exists() {
		for _, message := range messages.Array() {
			if message.Get("severity").String() == "error" {
				return fmt.Errorf("get error message %v", message.Get("string").String())
			}
		}
	}
	return nil
}

// device represent a disk found by smartctl
type device struct {
	Name     string `json:"name"`
	InfoName string `json:"info_name"`
	Type     string `json:"type"`
	Protocol string `json:"protocol"`
}

// ScanDevice scan all devices exist on machine
func ScanDevice() ([]device, error) {
	var (
		jsonResult gjson.Result
		err        error
		devices    []device
	)

	if jsonResult, err = getSMARTCtlResult("", "--scan"); err != nil {
		return nil, err
	}

	for _, result := range jsonResult.Get(_SMARTDevices).Array() {
		devices = append(devices, device{
			Name:     result.Get("name").String(),
			InfoName: result.Get("info_name").String(),
			Type:     result.Get("type").String(),
			Protocol: result.Get("protocol").String(),
		})
	}

	return devices, nil
}
