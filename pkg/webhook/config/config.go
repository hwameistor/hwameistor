package config

import (
	"bytes"
	"context"
	"github.com/hwameistor/hwameistor/pkg/local-disk-manager/utils/kubernetes"
	"github.com/hwameistor/hwameistor/pkg/utils"
	"github.com/hwameistor/hwameistor/pkg/utils/certmanager"
	log "github.com/sirupsen/logrus"
	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8s "k8s.io/client-go/kubernetes"
	"os"
	"path/filepath"
	"time"
)

var (
	// webhook config
	webhookNamespace, _ = os.LookupEnv("WEBHOOK_NAMESPACE")
	mutationCfgName, _  = os.LookupEnv("MUTATE_CONFIG")
	validateCfgName, _  = os.LookupEnv("VALIDATE_CONFIG")
	webhookService, _   = os.LookupEnv("WEBHOOK_SERVICE")
	validatePath, _     = os.LookupEnv("VALIDATE_PATH")
	mutationPath, _     = os.LookupEnv("MUTATE_PATH")
	dnsNames            = []string{webhookService, webhookService + "." + webhookNamespace, webhookService + "." + webhookNamespace + "." + "svc"}
	commonName          = webhookService + "." + webhookNamespace + "." + "svc"
	excludeNameSpaceKey = "name"

	// certs
	certsDir          = "/etc/webhook/certs"
	certKey           = "tls.key"
	certFile          = "tls.crt"
	Organization      = "hwameistor.io"
	DefaultEffecttime = 10
)

// CreateOrUpdateWebHookConfig create or update webhook config
func CreateOrUpdateWebHookConfig() error {
	// generate self-signed certs
	serverCertPEM, serverPrivateKeyPEM, err := certmanager.NewCertManager(
		[]string{Organization},
		time.Until(time.Date(time.Now().Year()+DefaultEffecttime, time.Now().Month(), time.Now().Day(), time.Now().Hour(), time.Now().Minute(), 0, 0, time.Now().Location())),
		dnsNames,
		commonName,
	).GenerateSelfSignedCerts()
	if err != nil {
		log.WithError(err).Error("failed to generate certs")
		return err
	}

	err = os.MkdirAll(certsDir, 0666)
	if err != nil {
		log.WithField("certDir", certsDir).WithError(err).Error("failed to create cert dir")
		return err
	}

	err = utils.WriteFile(filepath.Join(certsDir, certFile), serverCertPEM)
	if err != nil {
		log.WithField("tls.cert", serverCertPEM.String()).WithError(err).Error("failed to write tls.cert")
		return err
	}

	err = utils.WriteFile(filepath.Join(certsDir, certKey), serverPrivateKeyPEM)
	if err != nil {
		log.WithField("tls.key", serverPrivateKeyPEM.String()).WithError(err).Error("failed to write tls.key")
		return err
	}

	if err = CreateAdmissionConfig(serverCertPEM); err != nil {
		log.WithField("tls.cert", serverCertPEM.String()).WithError(err).Error("failed to create admission config")
	}

	return nil
}

func CreateAdmissionConfig(caCert *bytes.Buffer) error {
	clientset, err := kubernetes.NewClientSet()
	if err != nil {
		return err
	}

	if err = ensureNameSpaceKeyExist(clientset); err != nil {
		return err
	}

	ctx := context.Background()
	if mutationCfgName != "" {
		mutateConfig := &admissionregistrationv1.MutatingWebhookConfiguration{
			ObjectMeta: metav1.ObjectMeta{
				Name: mutationCfgName,
			},
			Webhooks: []admissionregistrationv1.MutatingWebhook{{
				Name: Organization + ".mutate-hook",
				ClientConfig: admissionregistrationv1.WebhookClientConfig{
					CABundle: caCert.Bytes(), // CA bundle created earlier
					Service: &admissionregistrationv1.ServiceReference{
						Name:      webhookService,
						Namespace: webhookNamespace,
						Path:      &mutationPath,
					},
				},
				Rules: []admissionregistrationv1.RuleWithOperations{{Operations: []admissionregistrationv1.OperationType{
					admissionregistrationv1.Create},
					Rule: admissionregistrationv1.Rule{
						APIGroups:   []string{"apps", ""},
						APIVersions: []string{"v1"},
						Resources:   []string{"pods"},
					},
				}},
				FailurePolicy: func() *admissionregistrationv1.FailurePolicyType {
					pt := admissionregistrationv1.Fail
					return &pt
				}(),
				NamespaceSelector: &metav1.LabelSelector{
					MatchExpressions: []metav1.LabelSelectorRequirement{
						{
							Key:      excludeNameSpaceKey,
							Operator: metav1.LabelSelectorOpNotIn,
							Values: []string{
								webhookNamespace,
							},
						},
					},
				},
				AdmissionReviewVersions: []string{"v1", "v1beta1"},
				SideEffects: func() *admissionregistrationv1.SideEffectClass {
					se := admissionregistrationv1.SideEffectClassNone
					return &se
				}(),
			}},
		}

		mutateAdmissionClient := clientset.AdmissionregistrationV1().MutatingWebhookConfigurations()
		m, err := mutateAdmissionClient.Get(ctx, mutationCfgName, metav1.GetOptions{})
		if err != nil {
			if errors.IsNotFound(err) {
				if _, err = mutateAdmissionClient.Create(ctx, mutateConfig, metav1.CreateOptions{}); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			mutateConfig.ResourceVersion = m.ResourceVersion
			if _, err = mutateAdmissionClient.Update(ctx, mutateConfig, metav1.UpdateOptions{}); err != nil {
				return err
			}
		}

	}

	return nil
}

func ensureNameSpaceKeyExist(clientset *k8s.Clientset) error {
	ns, err := clientset.CoreV1().Namespaces().Get(context.Background(), webhookNamespace, metav1.GetOptions{})
	if err != nil {
		return err
	}

	existLabels := ns.GetObjectMeta().GetLabels()
	if v, ok := existLabels[excludeNameSpaceKey]; ok && v == webhookNamespace {
		log.WithField(excludeNameSpaceKey, v).Debug("namespace selector is exited")
		return nil
	}
	existLabels[excludeNameSpaceKey] = webhookNamespace
	ns.ObjectMeta.Labels = existLabels

	// update namespace labels
	_, err = clientset.CoreV1().Namespaces().Update(context.Background(), ns, metav1.UpdateOptions{})
	if err != nil {
		log.WithError(err).WithFields(log.Fields{
			"namespace": ns.Name,
			"labels":    existLabels,
		}).Error("failed to update namespace labels")
		return err
	}

	return nil
}
