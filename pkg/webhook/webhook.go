package webhook

import (
	"github.com/spf13/pflag"
	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
)

var (
	PodResource           = metav1.GroupVersionResource{Version: "v1", Resource: "pods"}
	MutateWebhooks        []MutateAdmissionWebhook
	UniversalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
)

func AddToMutateHooks(hook MutateAdmissionWebhook) {
	MutateWebhooks = append(MutateWebhooks, hook)
}

// PatchOperation used for mutate webhook
type PatchOperation struct {
	Operation string      `json:"op"`
	Path      string      `json:"path"`
	Value     interface{} `json:"value,omitempty"`
}

type AdmissionWebhook interface {
	ResourceNeedHandle(admission.AdmissionReview) (bool, error)
	Name() string
	Init(ServerOption)
}
type MutateAdmissionWebhook interface {
	AdmissionWebhook

	// Mutate resources
	Mutate(admission.AdmissionReview) ([]PatchOperation, error)
}

type ValidateAdmissionWebhook interface {
	AdmissionWebhook

	Validate() interface{}
}

type ServerOption struct {
	SchedulerName string `json:"scheduler_name"`
	CertDir       string `json:"cert_dir"`
	TLSCert       string `json:"tls_cert"`
	TLSKey        string `json:"tls_key"`
}

func NewServerOption() *ServerOption {
	return &ServerOption{}
}

func (s *ServerOption) AddFlags(fs *pflag.FlagSet) *pflag.FlagSet {
	fs.StringVar(&s.CertDir, "cert-dir", s.CertDir, "The directory where the TLS certs are located. "+
		"If --tls-cert-file and --tls-private-key-file are provided, this flag will be ignored.")

	fs.StringVar(&s.TLSCert, "tls-cert-file", s.TLSCert, ""+
		"File containing the default x509 Certificate for HTTPS. (CA cert, if any, concatenated "+
		"after server cert). If HTTPS serving is enabled, and --tls-cert-file and "+
		"--tls-private-key-file are not provided, a self-signed certificate and key "+
		"are generated for the public address and saved to the directory specified by --cert-dir.")

	fs.StringVar(&s.TLSKey, "tls-private-key-file", s.TLSKey,
		"File containing the default x509 private key matching --tls-cert-file.")

	fs.StringVar(&s.SchedulerName, "scheduler-name", s.SchedulerName,
		"Scheduler name is used to compare with the scheduler name in pod.spec.schedulerName, "+
			"you can find which scheduler name hwameistor is now used in ConfigMap/hwameistor-scheduler-config, "+
			"the default value is hwameistor-scheduler.")
	return fs
}
