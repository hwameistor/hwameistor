package framework

import (
	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	ctrlclient "sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/apiutil"
	"sigs.k8s.io/controller-runtime/pkg/client/config"
	"time"
)

// Framework supports common operations used by e2e tests; it will keep a client for you.
type Framework struct {
	scheme           *runtime.Scheme
	addToSchemeFuncs []func(s *runtime.Scheme) error
	clientConfig     *rest.Config
	client           ctrlclient.Client
	options          Options
	timeouts         *TimeoutContext
}

// GetClient returns controller runtime client
func (f *Framework) GetClient() ctrlclient.Client {
	return f.client
}

// GetScheme returns scheme
func (f *Framework) GetScheme() *runtime.Scheme {
	return f.scheme
}

// GetTimeouts returns timeouts
func (f *Framework) GetTimeouts() *TimeoutContext {
	return f.timeouts
}

// ClientConfig returns rest configs
func (f *Framework) ClientConfig() *rest.Config {
	return f.clientConfig
}

// Options is a struct for managing test framework options.
type Options struct {
	ClientQPS   float32
	ClientBurst int
}

// NewFrameworkWithCustomTimeouts makes a framework with with custom timeouts.
func NewFrameworkWithCustomTimeouts(timeouts *TimeoutContext, addToSchemeFuncs ...func(s *runtime.Scheme) error) *Framework {
	f := NewDefaultFramework(addToSchemeFuncs...)
	//	f.timeouts = timeouts
	return f
}

// NewDefaultFramework makes a new framework and sets up clientsets.
func NewDefaultFramework(addToSchemeFuncs ...func(s *runtime.Scheme) error) *Framework {
	options := Options{
		ClientQPS:   20,
		ClientBurst: 50,
	}
	f := NewFramework(options, nil, addToSchemeFuncs...)
	f.defaultConfig()
	return f
}

// NewFramework creates a test framework.
func NewFramework(options Options, client ctrlclient.Client, addToSchemeFuncs ...func(s *runtime.Scheme) error) *Framework {
	f := &Framework{
		options:          options,
		addToSchemeFuncs: addToSchemeFuncs,
		client:           client,
		timeouts:         NewTimeoutContextWithDefaults(),
	}

	return f
}

// DefaultBeforeEach gets clientsets
func (f *Framework) defaultConfig() {
	if f.client == nil {

		cfg, err := config.GetConfig()
		f.ExpectNoError(err)

		cfg.QPS = f.options.ClientQPS
		cfg.Burst = f.options.ClientBurst

		// Create the mapper provider
		mapper, err := apiutil.NewDynamicRESTMapper(cfg)
		err = wait.PollImmediate(5*time.Second, 3*time.Minute, func() (done bool, err error) {
			_, err = apiutil.NewDynamicRESTMapper(cfg)
			if err != nil {
				return false, nil
			}
			return true, nil
		})
		mapper, err = apiutil.NewDynamicRESTMapper(cfg)
		if err != nil {
			f.ExpectNoError(err)
		}
		sche := scheme.Scheme
		if len(f.addToSchemeFuncs) > 0 {
			for _, fn := range f.addToSchemeFuncs {
				if err := fn(sche); err != nil {
					f.ExpectNoError(err)
				}
			}
		}
		client, err := ctrlclient.New(cfg, ctrlclient.Options{Scheme: sche, Mapper: mapper})
		if err != nil {
			f.ExpectNoError(err)
		}

		f.client = client
	}
}

// ExpectEqual expects the specified two are the same, otherwise an exception raises
func (f *Framework) ExpectEqual(actual interface{}, extra interface{}, explain ...interface{}) {
	gomega.ExpectWithOffset(1, actual).To(gomega.Equal(extra), explain...)
}

// ExpectNotEqual expects the specified two are not the same, otherwise an exception raises
func (f *Framework) ExpectNotEqual(actual interface{}, extra interface{}, explain ...interface{}) {
	gomega.ExpectWithOffset(1, actual).NotTo(gomega.Equal(extra), explain...)
}

// ExpectError expects an error happens, otherwise an exception raises
func (f *Framework) ExpectError(err error, explain ...interface{}) {
	gomega.ExpectWithOffset(1, err).To(gomega.HaveOccurred(), explain...)
}

// ExpectNoError checks if "err" is set, and if so, fails assertion while logging the error.
func (f *Framework) ExpectNoError(err error, explain ...interface{}) {
	f.ExpectNoErrorWithOffset(1, err, explain...)
}

// ExpectNoErrorWithOffset checks if "err" is set, and if so, fails assertion while logging the error at "offset" levels above its caller
// (for example, for call chain f -> g -> ExpectNoErrorWithOffset(1, ...) error would be logged for "f").
func (f *Framework) ExpectNoErrorWithOffset(offset int, err error, explain ...interface{}) {
	gomega.ExpectWithOffset(1+offset, err).NotTo(gomega.HaveOccurred(), explain...)
}

// ExpectConsistOf expects actual contains precisely the extra elements.  The ordering of the elements does not matter.
func (f *Framework) ExpectConsistOf(actual interface{}, extra interface{}, explain ...interface{}) {
	gomega.ExpectWithOffset(1, actual).To(gomega.ConsistOf(extra), explain...)
}

// ExpectHaveKey expects the actual map has the key in the keyset
func (f *Framework) ExpectHaveKey(actual interface{}, key interface{}, explain ...interface{}) {
	gomega.ExpectWithOffset(1, actual).To(gomega.HaveKey(key), explain...)
}

// ExpectEmpty expects actual is empty
func (f *Framework) ExpectEmpty(actual interface{}, explain ...interface{}) {
	gomega.ExpectWithOffset(1, actual).To(gomega.BeEmpty(), explain...)
}
