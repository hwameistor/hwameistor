package app

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/hwameistor/hwameistor/pkg/webhook"
	hookcfg "github.com/hwameistor/hwameistor/pkg/webhook/config"
	"github.com/hwameistor/hwameistor/pkg/webhook/scheduler"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
	admission "k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"net/http"
)

const (
	jsonContentType = `application/json`
)

func doServerHwameiStorMutateFunc(w http.ResponseWriter, r *http.Request, o webhook.ServerOption, hooks ...webhook.MutateAdmissionWebhook) ([]byte, error) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, fmt.Errorf("invalid method %s, only POST requests are allowed", r.Method)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("could not read request body: %v", err)
	}

	if contentType := r.Header.Get("Content-Type"); contentType != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("unsupported content type %s, only %s is supported", contentType, jsonContentType)
	}

	// Parse the AdmissionReview request.
	var admissionReviewReq admission.AdmissionReview

	if _, _, err := webhook.UniversalDeserializer.Decode(body, nil, &admissionReviewReq); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("could not deserialize request: %v", err)
	} else if admissionReviewReq.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("malformed admission review: request is nil")
	}

	logCtx := log.Fields{
		"Group":     admissionReviewReq.Request.Resource.Group,
		"Kind":      admissionReviewReq.Request.Kind.Kind,
		"Version":   admissionReviewReq.Request.Resource.Version,
		"Namespace": admissionReviewReq.Request.Namespace,
		"Name":      admissionReviewReq.Request.Name,
	}

	// Do mutate hooks and construct the AdmissionReview response.
	admissionReviewResponse := admission.AdmissionReview{
		TypeMeta: admissionReviewReq.TypeMeta,
		Response: &admission.AdmissionResponse{
			UID: admissionReviewReq.Request.UID,
		},
	}

	var patchOptions []webhook.PatchOperation
	complete := false
	for _, hook := range hooks {
		// fixme: move init at some other place
		hook.Init(o)
		need, err := hook.ResourceNeedHandle(admissionReviewReq)
		if err != nil {
			log.WithFields(logCtx).Errorf("webhook %s failed to judge request if need handle", hook.Name())
			admissionReviewResponse.Response.Allowed = false
			admissionReviewResponse.Response.Result = &metav1.Status{
				Message: err.Error(),
			}
			complete = true
			break
		}

		if !need {
			log.WithFields(logCtx).Debugf("skip mutate webhooks %s", hook.Name())
			continue
		}

		ops, err := hook.Mutate(admissionReviewReq)
		if err != nil {
			log.WithFields(logCtx).Errorf("webhook %s failed to mutate request", hook.Name())
			admissionReviewResponse.Response.Allowed = false
			admissionReviewResponse.Response.Result = &metav1.Status{
				Message: err.Error(),
			}
			complete = true
			break
		}
		patchOptions = append(patchOptions, ops...)
	}

	if !complete {
		// Otherwise, encode the patch operations to JSON and return a positive response.
		patchBytes, err := json.Marshal(patchOptions)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return nil, fmt.Errorf("could not marshal JSON patch: %v", err)
		}
		admissionReviewResponse.Response.Allowed = true
		admissionReviewResponse.Response.Patch = patchBytes
		admissionReviewResponse.Response.PatchType = new(admission.PatchType)
		*admissionReviewResponse.Response.PatchType = admission.PatchTypeJSONPatch
		complete = true
	}

	bytes, err := json.Marshal(&admissionReviewResponse)
	if err != nil {
		return nil, fmt.Errorf("marshaling response: %v", err)
	}
	return bytes, nil
}

func serverHwameiStorMutateFunc(w http.ResponseWriter, r *http.Request, o webhook.ServerOption) {
	logCtx := log.Fields{"request.path": r.URL.Path}

	var writeErr error
	if bytes, err := doServerHwameiStorMutateFunc(w, r, o, webhook.MutateWebhooks...); err != nil {
		log.WithFields(logCtx).WithError(err).Error("failed to handle mutate webhook request")
		w.WriteHeader(http.StatusInternalServerError)
		_, writeErr = w.Write([]byte(err.Error()))
	} else {
		log.WithFields(logCtx).Debug("handle mutate webhook successfully")
		_, writeErr = w.Write(bytes)
	}

	if writeErr != nil {
		log.WithFields(logCtx).WithError(writeErr).Error("failed to write response")
	}
}

func RegisterHwameiStorMutateWebhooks(o webhook.ServerOption) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		serverHwameiStorMutateFunc(w, r, o)
	})
}

func init() {
	webhook.AddToMutateHooks(scheduler.NewPatchSchedulerWebHook())
	if err := hookcfg.CreateOrUpdateWebHookConfig(); err != nil {
		panic(err)
	}
}
