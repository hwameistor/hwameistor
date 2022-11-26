package smart

import (
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils"
	log "github.com/sirupsen/logrus"
	"github.com/tidwall/gjson"
)

const (
	_SMARTCtl = "smartctl"

	// json path in smartctl result
	_SMARTExitStatus   = "smartctl.exit_status"
	_SMARTMessages     = "smartctl.messages"
	_SMARTStatusPassed = "smart_status.passed"
)

type Device struct {
	Device string
	Serial string
	Model  string
}

type controller struct {
	Device  Device
	Options []string
}

// NewSMARTController used to get S.M.A.R.T info
func NewSMARTController(device Device, options ...string) *controller {
	return &controller{
		Device:  device,
		Options: options,
	}
}

// SupportSmart judge if this device supports SMART or not
func (c *controller) SupportSmart() (bool, error) {
	var (
		jsonStatus gjson.Result
		err        error
	)

	if jsonStatus, err = getSMARTCtl(c.Device.Device, append(c.Options, "-i")...); err != nil {
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

	if jsonStatus, err = getSMARTCtl(c.Device.Device, append(c.Options, "-H")...); err != nil {
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

	if jsonStatus, err = getSMARTCtl(c.Device.Device, append(c.Options, "/c0", "show")...); err != nil {
		return false, err
	}

	return jsonStatus.Get(_SMARTStatusPassed).Bool(), err
}

// getSMARTCtl get smartctl output
func getSMARTCtl(device string, options ...string) (gjson.Result, error) {
	var (
		result gjson.Result
	)

	out, err := utils.BashWithArgs(_SMARTCtl, append(options, device, "--json")...)
	if err != nil {
		log.WithError(err).Error(out)
		return result, err
	}
	if !gjson.Valid(out) {
		log.Error("invalid json format: %v", out)
		return result, fmt.Errorf("invalid json format")
	}

	result = gjson.Parse(out)
	if err = resultCodeIsOk(result.Get(_SMARTExitStatus).Int()); err != nil {
		log.WithError(err).Error(out)
		return result, err
	}

	if err = jsonIsOk(result); err != nil {
		log.WithError(err).Error("got error message")
		return result, err
	}

	return result, nil
}

// Parse smartctl return code
func resultCodeIsOk(SMARTCtlResult int64) error {
	var (
		err error
	)
	if SMARTCtlResult > 0 {
		b := SMARTCtlResult
		if (b & 1) != 0 {
			err = fmt.Errorf("command line did not parse")
		}

		if (b & (1 << 3)) != 0 {
			err = fmt.Errorf("SMART status check returned 'DISK FAILING'")
		}
		if (b & (1 << 5)) != 0 {
			err = fmt.Errorf("SMART status check returned 'DISK OK' but we found that " +
				"some (usage or prefail) Attributes have been <= threshold at some time in the past")
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
