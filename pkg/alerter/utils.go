package alerter

import (
	"context"
	"fmt"
	"time"

	localstoragev1alpha1 "github.com/hwameistor/local-storage/pkg/apis/localstorage/v1alpha1"
	log "github.com/sirupsen/logrus"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func genTimeStampString() string {
	t := time.Now().Local()
	return fmt.Sprintf("%d-%d-%dt%d-%d-%d", t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second())
}

func createAlert(alert *localstoragev1alpha1.LocalStorageAlert) {
	log.WithFields(log.Fields{
		"alert":    alert.Name,
		"severity": alert.Spec.Severity,
		"module":   alert.Spec.Module,
		"resource": alert.Spec.Resource,
	}).Debug("Generating an alert")
	if _, err := gAlertClient.Create(context.TODO(), alert, metav1.CreateOptions{}); err != nil {
		log.WithField("name", alert.Name).WithError(err).Error("Failed to create an alert CRD")
	}
}
