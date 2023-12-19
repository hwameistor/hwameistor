package faultmanagement

import (
	"github.com/hwameistor/hwameistor/pkg/apis/hwameistor/v1alpha1"
	log "github.com/sirupsen/logrus"
)

// processFaultTicketEvaluating a new faultTicket that should be evaluated first, it has the following steps:
// 1. Find out which fault type it is
// 2. List all the resources affected by this fault according to Resource Relation Map
// 3. Update the above info to the Status field
func (m *manager) processFaultTicketEvaluating(faultTicket *v1alpha1.FaultTicket) error {
	logger := m.logger.WithFields(log.Fields{
		"faultTicket": faultTicket.Name,
		"faultType":   faultTicket.Spec.Type,
		"node":        faultTicket.Spec.NodeName,
		"source":      faultTicket.Spec.Source,
		"message":     faultTicket.Spec.Message,
	})
	logger.Debug("Starting faultTicket evaluation")

	// TODO(ming): handler these fault according to the config that admin applied
	var err error
	switch faultTicket.Spec.Type {
	case v1alpha1.DiskFaultTicket:
		err = m.evaluatingDiskFault(faultTicket)

	case v1alpha1.VolumeFaultTicket:
		err = m.evaluatingVolumeFault(faultTicket)

	case v1alpha1.NodeFaultTicket:
		err = m.evaluatingNodeFault(faultTicket)

	default:
		logger.Debug("Unknown Fault Type, ignore it")
	}

	return err
}

func (m *manager) evaluatingDiskFault(faultTicket *v1alpha1.FaultTicket) error {
	// m.topologyGraph.Draw()
	return nil
}

func (m *manager) evaluatingVolumeFault(faultTicket *v1alpha1.FaultTicket) error {
	return nil
}

func (m *manager) evaluatingNodeFault(faultTicket *v1alpha1.FaultTicket) error {
	return nil
}
