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
	"fmt"
	"sort"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	labels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	kupidv1alpha1 "github.com/gardener/kupid/api/v1alpha1"
	"github.com/gardener/kupid/pkg/common"
	"github.com/go-logr/logr"
)

type processorFactory interface {
	kind() metav1.GroupVersionKind
	isMutating() bool
	newProcessor(logr.Logger) processor
}

type objectGetter interface {
	getObject() runtime.Object
}

type processor interface {
	objectGetter
	process(ctx context.Context, cacheReader, directReader client.Reader, namespace string) (mutated bool, err error)
}

// podSpecCallbacks defines the callback necessary to implement the processor to mutate pod specs embedded in supported objects.
type podSpecCallbacks interface {
	objectGetter
	getPodSpec() *corev1.PodSpec
	getPodLabels() map[string]string
}

type injectSchedulingPolicyInjector interface {
	injectSchedulingPolicyInjector(injector schedulingPolicyInjector)
}

// podSpecProcessorImpl implements the processor interface based on the supplied podSpecCallbacks and schedulingPolicyInjector.
type podSpecProcessorImpl struct {
	podSpecCallbacks
	injector schedulingPolicyInjector
	logger   logr.Logger
}

func (p *podSpecProcessorImpl) injectSchedulingPolicyInjector(injector schedulingPolicyInjector) {
	p.injector = injector
}

func (p *podSpecProcessorImpl) process(ctx context.Context, cacheReader, directReader client.Reader, namespace string) (bool, error) {
	var (
		obj       = p.getObject()
		podSpec   = p.getPodSpec()
		podLabels = p.getPodLabels()
		l         = p.logger
	)

	l.V(1).Info("Beginning of mutate", "obj", obj)

	csps, err := p.getAllClusterSchedulingPolicies(ctx, cacheReader)
	l.V(1).Info("Getting ClusterSchedulingPolicies", "csps", csps, "Error", err)
	if err != nil {
		l.V(1).Info("Error getting ClusterSchedulingPolicies from cache", "csps", csps, "Error:", err)
		csps, err = p.getAllClusterSchedulingPolicies(ctx, directReader)
		if err != nil {
			l.V(1).Info("Error getting ClusterSchedulingPolicies with direct reader", "csps", csps, "Error:", err)
			return false, err
		}
	}

	ns, err := p.getNamespace(ctx, cacheReader, namespace)
	l.V(1).Info("Getting namespace from cache", "ns", ns, "Error", err)
	if err != nil {
		l.V(1).Info("Error getting namespace from cache", "ns", ns, "Error", err)
		ns, err = p.getNamespace(ctx, directReader, namespace)
		if err != nil {
			l.V(1).Info("Error getting namespace from direct reader", "ns", ns, "Error", err)
			return false, err
		}
	}

	sps, err := p.getAllSchedulingPoliciesInNamespace(ctx, cacheReader, namespace)
	l.V(1).Info("Getting SchedulingPolicies from Cache", "sps", sps, "Error", err)
	if err != nil {
		l.V(1).Info("Error getting SchedulingPolicies from cache", "sps", sps, "Error", err)
		sps, err = p.getAllSchedulingPoliciesInNamespace(ctx, directReader, namespace)
		if err != nil {
			l.V(1).Info("Error getting SchedulingPolicies with direct reader", "sps", sps, "Error", err)
			return false, err
		}
	}

	var spcs []common.PodSchedulingPolicyConfiguration

	if filtered, err := p.filterClusterSchedulingPolicies(csps, ns, podLabels); err != nil {
		return false, err
	} else if len(filtered) > 0 {
		spcs = append(spcs, filtered...)
	}

	if filtered, err := p.filterSchedulingPolicies(sps, podLabels); err != nil {
		return false, err
	} else if len(filtered) > 0 {
		spcs = append(spcs, filtered...)
	}

	sort.Slice(spcs, func(i, j int) bool {
		if a, err := meta.Accessor(spcs[i]); err != nil {
			panic(err)
		} else if b, err := meta.Accessor(spcs[i]); err != nil {
			panic(err)
		} else {
			return a.GetName() < b.GetName()
		}
	})

	l.V(1).Info("Applicable scheduling policy configuration", "spcs", spcs)
	if len(spcs) <= 0 {
		return false, nil
	}

	if err := p.injectSchedulingPolicies(spcs, podSpec); err != nil {
		return false, err
	}

	return true, nil
}

func (p *podSpecProcessorImpl) getAllClusterSchedulingPolicies(ctx context.Context, reader client.Reader) ([]kupidv1alpha1.ClusterPodSchedulingPolicy, error) {
	var (
		items         []kupidv1alpha1.ClusterPodSchedulingPolicy
		continueToken string
	)

	for {
		list := &kupidv1alpha1.ClusterPodSchedulingPolicyList{}
		if err := reader.List(ctx, list, client.Continue(continueToken)); err != nil {
			return items, err
		}

		items = append(items, list.Items...)
		continueToken := list.Continue
		if continueToken == "" {
			return items, nil
		}
	}
}

func (p *podSpecProcessorImpl) getAllSchedulingPoliciesInNamespace(ctx context.Context, reader client.Reader, nsName string) ([]kupidv1alpha1.PodSchedulingPolicy, error) {
	var (
		items         []kupidv1alpha1.PodSchedulingPolicy
		continueToken string
	)

	for {
		list := &kupidv1alpha1.PodSchedulingPolicyList{}
		if err := reader.List(ctx, list, client.InNamespace(nsName), client.Continue(continueToken)); err != nil {
			return items, err
		}

		items = append(items, list.Items...)
		continueToken := list.Continue
		if continueToken == "" {
			return items, nil
		}
	}
}

func (p *podSpecProcessorImpl) getNamespace(ctx context.Context, reader client.Reader, nsName string) (*corev1.Namespace, error) {
	var ns = &corev1.Namespace{}
	if err := reader.Get(ctx, client.ObjectKey{Name: nsName}, ns); err != nil {
		return ns, err
	}

	return ns, nil
}

func (p *podSpecProcessorImpl) filterClusterSchedulingPolicies(csps []kupidv1alpha1.ClusterPodSchedulingPolicy, ns *corev1.Namespace, podLabels map[string]string) ([]common.PodSchedulingPolicyConfiguration, error) {
	var (
		filtered []common.PodSchedulingPolicyConfiguration
		l        = p.logger
	)

	for i := range csps {
		csp := &csps[i]

		if s, err := metav1.LabelSelectorAsSelector(csp.Spec.NamespaceSelector); err != nil {
			return filtered, err
		} else if !s.Matches(labels.Set(ns.Labels)) {
			l.V(1).Info("namespaceSelector match failed", "selector", s, "labels", ns.Labels)
			continue
		}

		if s, err := metav1.LabelSelectorAsSelector(csp.Spec.PodSelector); err != nil {
			return filtered, err
		} else if s.Matches(labels.Set(podLabels)) {
			filtered = append(filtered, csp)
		} else {
			l.V(1).Info("podSelector match failed", "selector", s, "labels", podLabels)
		}
	}

	return filtered, nil
}

func (p *podSpecProcessorImpl) filterSchedulingPolicies(sps []kupidv1alpha1.PodSchedulingPolicy, podLabels map[string]string) ([]common.PodSchedulingPolicyConfiguration, error) {
	var filtered []common.PodSchedulingPolicyConfiguration

	for i := range sps {
		sp := &sps[i]

		if s, err := metav1.LabelSelectorAsSelector(sp.Spec.PodSelector); err != nil {
			return filtered, err
		} else if s.Matches(labels.Set(podLabels)) {
			filtered = append(filtered, sp)
		}
	}

	return filtered, nil
}

func (p *podSpecProcessorImpl) injectSchedulingPolicies(spcs []common.PodSchedulingPolicyConfiguration, podSpec *corev1.PodSpec) error {
	if p.injector == nil {
		return fmt.Errorf("schedulingPolicyInjector should be injected before handling any requests")
	}
	orig := podSpec.DeepCopy()
	for _, spc := range spcs {
		p.injector.injectPodSchedulingPolicyConfiguration(spc, orig, podSpec)
	}
	return nil
}
