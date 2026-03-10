package config

import (
	"testing"

	admissionregistrationv1 "k8s.io/api/admissionregistration/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestMergeWebhookPreservesNamespaceSelectorButRefreshesFailurePolicy(t *testing.T) {
	ignore := admissionregistrationv1.Ignore
	fail := admissionregistrationv1.Fail

	update := &admissionregistrationv1.MutatingWebhookConfiguration{
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			FailurePolicy: &ignore,
			NamespaceSelector: &metav1.LabelSelector{
				MatchExpressions: []metav1.LabelSelectorRequirement{{
					Key:      "hwameistor.io/webhook",
					Operator: metav1.LabelSelectorOpNotIn,
					Values:   []string{"ignore"},
				}},
			},
		}},
	}

	existing := &admissionregistrationv1.MutatingWebhookConfiguration{
		Webhooks: []admissionregistrationv1.MutatingWebhook{{
			FailurePolicy: &fail,
			NamespaceSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"custom": "selector"},
			},
		}},
	}

	mergeExistingWebhookSettings(update, existing)

	if update.Webhooks[0].FailurePolicy == nil || *update.Webhooks[0].FailurePolicy != ignore {
		t.Fatalf("expected failure policy to remain %q, got %#v", ignore, update.Webhooks[0].FailurePolicy)
	}

	if got := update.Webhooks[0].NamespaceSelector; got == nil || got.MatchLabels["custom"] != "selector" {
		t.Fatalf("expected namespace selector to be preserved, got %#v", got)
	}
}
