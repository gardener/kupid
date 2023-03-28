// Copyright (c) 2020 SAP SE or an SAP affiliate company. All rights reserved. This file is licensed under the Apache Software License, v. 2 except as noted otherwise in the LICENSE file.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package webhook

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	utilruntime "k8s.io/apimachinery/pkg/util/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/cache"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/metrics"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"

	kupidv1alpha1 "github.com/gardener/kupid/api/v1alpha1"
	"github.com/prometheus/client_golang/prometheus"
)

// +kubebuilder:rbac:groups=kupid.gardener.cloud,resources=clusterpodschedulingpolicies;podschedulingpolicies,verbs=get;list;watch;create;update;patch;delete

// WebhookPath is the path under which the webhook will be registered.
const WebhookPath = "/webhook"

var log = ctrl.Log.WithName("webhook")

// Webhook is the implementation for all the webhooks required for kupid.
type Webhook struct {
	cache              cache.Cache
	directReader       client.Reader
	mu                 sync.Mutex
	once               sync.Once
	injector           schedulingPolicyInjector
	decoder            *admission.Decoder
	started            bool
	processorFactories map[metav1.GroupVersionKind]processorFactory
}

// NewWebhook creates a bare Webhook instance.
func NewWebhook() *Webhook {
	return &Webhook{
		processorFactories: map[metav1.GroupVersionKind]processorFactory{},
	}
}

// NewWebhookWithProcessorFactories creates a Webhook instance with the supplied processor factories registered.
func NewWebhookWithProcessorFactories(pfs []processorFactory) (*Webhook, error) {
	var w = NewWebhook()
	return w, w.registerProcessorFactories(pfs)
}

// NewDefaultWebhook creates a Webhook instance with the default eet of processor factories registered.
func NewDefaultWebhook() (*Webhook, error) {
	return NewWebhookWithProcessorFactories([]processorFactory{
		&clusterPodSchedulingPolicyProcessorFactory{},
		&podSchedulingPolicyProcessorFactory{},
		&podProcessorFactory{},
		&replicationControllerProcessorFactory{},
		&replicaSetProcessorFactory{},
		&deploymentProcessorFactory{},
		&statefulSetProcessorFactory{},
		&daemonSetProcessorFactory{},
		&jobProcessorFactory{},
		&cronJobProcessorFactory{},
	})
}

func (w *Webhook) registerProcessorFactory(pf processorFactory) error {
	var gvk = pf.kind()
	log.Info("Registering processor factory", "kind", gvk)
	if _, ok := w.processorFactories[gvk]; ok {
		return fmt.Errorf("re-registering processor factory is not allowed for kind: %s", gvk)
	}
	w.processorFactories[gvk] = pf
	return nil
}

func (w *Webhook) registerProcessorFactories(pfs []processorFactory) error {
	for _, pf := range pfs {
		if err := w.registerProcessorFactory(pf); err != nil {
			return err
		}
	}
	return nil
}

func (w *Webhook) getProcessorFactoryFor(gvk metav1.GroupVersionKind) (processorFactory, error) {
	pf, ok := w.processorFactories[gvk]
	if !ok {
		return pf, fmt.Errorf("no processor factory registered for kind: %s", gvk)
	}
	return pf, nil
}

// InjectCache injects a cache into the webhook.
func (w *Webhook) InjectCache(c cache.Cache) error {
	w.cache = c
	return nil
}

// InjectAPIReader injects a direct client.Reader into the webhook.
func (w *Webhook) InjectAPIReader(r client.Reader) error {
	w.directReader = r
	return nil
}

func (w *Webhook) waitForCacheSyncOnce(ctx context.Context) {
	w.once.Do(func() {
		log.Info("Waiting for caches to be synced")
		if ok := w.cache.WaitForCacheSync(ctx); !ok {
			err := fmt.Errorf("failed to wait for caches to sync")
			log.Error(err, "Could not wait for Cache to sync")
		}

		log.Info("Cache initialized")
	})
}

// Start initializes the cache.  The component will stop running
// when the channel is closed.  Start blocks until the channel is closed or
// an error occurs.
func (w *Webhook) Start(ctx context.Context) error {
	// use an IIFE to get proper lock handling
	// but lock outside to get proper handling of the queue shutdown
	w.mu.Lock()

	err := func() error {
		defer w.mu.Unlock()

		// TODO(pwittrock): Reconsider HandleCrash
		defer utilruntime.HandleCrash()

		// cache should have been injected before Start was called
		if w.cache == nil {
			return fmt.Errorf("must call InjectCache on Runnable before calling Start")
		}

		// directReader should have been injected before Start was called
		if w.directReader == nil {
			return fmt.Errorf("must call InjectDirectReader on Runnable before calling Start")
		}

		for _, obj := range []client.Object{
			&corev1.Namespace{},
			&kupidv1alpha1.ClusterPodSchedulingPolicy{},
			&kupidv1alpha1.PodSchedulingPolicy{},
		} {
			log.Info("Registering informer", "obj", obj)
			if _, err := w.cache.GetInformer(ctx, obj); err != nil {
				if kindMatchErr, ok := err.(*meta.NoKindMatchError); ok {
					log.Error(err, "if kind is a CRD, it should be installed before calling Start",
						"kind", kindMatchErr.GroupKind)
				}
				return err
			}
		}

		w.waitForCacheSyncOnce(ctx)

		w.started = true
		return nil
	}()
	if err != nil {
		return err
	}

	return nil
}

// SetupWithManager sets up the webhook to inject applicable podschedulingpolicies and clusterpodschedulingpolicies.
func (w *Webhook) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.Add(w); err != nil {
		return err
	}

	mgr.GetWebhookServer().Register(WebhookPath, &admission.Webhook{Handler: w})
	log.Info("Webhook registered")

	return mgr.AddHealthzCheck("webhook", func(req *http.Request) error { return nil })
}

// InjectDecoder injects the decoder into a webhook.
func (w *Webhook) InjectDecoder(d *admission.Decoder) error {
	w.decoder = d
	return nil
}

// Handle handles admission requests.
func (w *Webhook) Handle(ctx context.Context, req admission.Request) admission.Response {
	l := log.WithValues("resource", req.Resource, "namespace", req.Namespace, "name", req.Name)

	l.V(1).Info("Handling", "req", req)
	l.Info("Handling request")

	w.waitForCacheSyncOnce(ctx)

	kupidRequestsTotal.With(prometheus.Labels{labelType: typeProcessed}).Inc()

	// Get the namespace in the request
	namespace := req.Namespace

	pf, err := w.getProcessorFactoryFor(req.Kind)
	if err != nil {
		l.Error(err, "Error getting processor factory")
		kupidRequestsTotal.With(prometheus.Labels{labelType: typeError}).Inc()
		return admission.Allowed(err.Error())
	}

	p := pf.newProcessor(l)

	if into, ok := p.(injectSchedulingPolicyInjector); ok {
		into.injectSchedulingPolicyInjector(w.getInjector())
	}

	obj := p.getObject()

	err = w.decoder.Decode(req, obj)
	if err != nil {
		l.Error(err, "Error decoding admission request object")
		kupidRequestsTotal.With(prometheus.Labels{labelType: typeError}).Inc()
		return admission.Allowed(err.Error())
	}

	if !pf.isMutating() {
		if _, err := p.process(ctx, w.cache, w.directReader, namespace); err != nil {
			l.Error(err, "Error processing admission request")
			kupidRequestsTotal.With(prometheus.Labels{labelType: typeDenied}).Inc()
			return admission.Denied(err.Error())
		}

		kupidRequestsTotal.With(prometheus.Labels{labelType: typeAllowed}).Inc()
		return admission.Allowed("")
	}

	if mutated, err := p.process(ctx, w.cache, w.directReader, namespace); err != nil {
		l.Error(err, "Error processing admission request")
		kupidRequestsTotal.With(prometheus.Labels{labelType: typeError}).Inc()
		return admission.Allowed(err.Error())
	} else if !mutated {
		l.Info("Nothing mutated for request")
		kupidRequestsTotal.With(prometheus.Labels{labelType: typeAllowed}).Inc()
		return admission.Allowed("")
	}

	obj = p.getObject()
	marshalled, err := json.Marshal(obj)
	if err != nil {
		l.Error(err, "Error marshalling mutated object for admission response")
		kupidRequestsTotal.With(prometheus.Labels{labelType: typeError}).Inc()
		return admission.Allowed(err.Error())
	}

	// Create the patch
	res := admission.PatchResponseFromRaw(req.Object.Raw, marshalled)
	if len(res.Patches) > 0 {
		l.V(1).Info("Mutated response", "res", res)
		l.Info("Mutated response for request")
		kupidRequestsTotal.With(prometheus.Labels{labelType: typeAllowed}).Inc()
		kupidRequestsTotal.With(prometheus.Labels{labelType: typeMutated}).Inc()
	}

	return res
}

func (w *Webhook) getInjector() schedulingPolicyInjector {
	if w.injector == nil {
		w.injector = newDefaultPodSchedulingPolicyInjector()
	}

	return w.injector
}

const (
	kupidNamespace     = "kupid"
	subsystemAggregate = "aggr"
	labelType          = "type"
	typeProcessed      = "processed"
	typeAllowed        = "allowed"
	typeDenied         = "denied"
	typeMutated        = "mutated"
	typeError          = "error"
)

var (
	kupidRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Namespace: kupidNamespace,
			Subsystem: subsystemAggregate,
			Name:      "requests_total",
			Help:      "The accumulated total number of requests processed by kupid.",
		},
		[]string{labelType},
	)
)

func init() {
	for _, lt := range []string{typeProcessed, typeAllowed, typeDenied, typeMutated, typeError} {
		kupidRequestsTotal.With(prometheus.Labels{labelType: lt}).Add(0)
	}

	metrics.Registry.MustRegister(kupidRequestsTotal)
}
