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
	"reflect"

	"github.com/gardener/kupid/pkg/common"
	corev1 "k8s.io/api/core/v1"
)

type affinityInjector interface {
	injectAffinity(affinity *corev1.Affinity, orig, mutable *corev1.PodSpec)
}

type affinityInjectorFunc func(affinity *corev1.Affinity, orig, mutable *corev1.PodSpec)

func (f affinityInjectorFunc) injectAffinity(affinity *corev1.Affinity, orig, mutable *corev1.PodSpec) {
	if f == nil {
		return
	}

	f(affinity, orig, mutable)
}

type nodeNameInjector interface {
	injectNodeName(nodeName string, orig, mutable *corev1.PodSpec)
}

type nodeNameInjectorFunc func(nodeName string, orig, mutable *corev1.PodSpec)

func (f nodeNameInjectorFunc) injectNodeName(nodeName string, orig, mutable *corev1.PodSpec) {
	if f == nil {
		return
	}

	f(nodeName, orig, mutable)
}

type nodeSelectorInjector interface {
	injectNodeSelector(nodeSelector map[string]string, orig, mutable *corev1.PodSpec)
}

type nodeSelectorInjectorFunc func(nodeSelector map[string]string, orig, mutable *corev1.PodSpec)

func (f nodeSelectorInjectorFunc) injectNodeSelector(nodeSelector map[string]string, orig, mutable *corev1.PodSpec) {
	if f == nil {
		return
	}

	f(nodeSelector, orig, mutable)
}

type schedulerNameInjector interface {
	injectSchedulerName(schedulerName string, orig, mutable *corev1.PodSpec)
}

type schedulerNameInjectorFunc func(schedulerName string, orig, mutable *corev1.PodSpec)

func (f schedulerNameInjectorFunc) injectSchedulerName(schedulerName string, orig, mutable *corev1.PodSpec) {
	if f == nil {
		return
	}

	f(schedulerName, orig, mutable)
}

type tolerationsInjector interface {
	injectTolerations(tolerations []corev1.Toleration, orig, mutable *corev1.PodSpec)
}

type tolerationsInjectorFunc func(tolerations []corev1.Toleration, orig, mutable *corev1.PodSpec)

func (f tolerationsInjectorFunc) injectTolerations(tolerations []corev1.Toleration, orig, mutable *corev1.PodSpec) {
	if f == nil {
		return
	}

	f(tolerations, orig, mutable)
}

type schedulingPolicyInjector interface {
	injectPodSchedulingPolicyConfiguration(spc common.PodSchedulingPolicyConfiguration, orig, mutable *corev1.PodSpec)
}

// funcs implements all the injectors.
type funcs struct {
	affinityInjector
	nodeNameInjector
	nodeSelectorInjector
	schedulerNameInjector
	tolerationsInjector
}

func (f *funcs) injectAffinity(affinity *corev1.Affinity, orig, mutable *corev1.PodSpec) {
	if f == nil || f.affinityInjector == nil || affinity == nil || orig == nil || mutable == nil {
		return
	}

	f.affinityInjector.injectAffinity(affinity, orig, mutable)
}

func (f *funcs) injectNodeName(nodeName string, orig, mutable *corev1.PodSpec) {
	if f == nil || f.nodeNameInjector == nil || nodeName == "" || orig == nil || mutable == nil {
		return
	}

	f.nodeNameInjector.injectNodeName(nodeName, orig, mutable)
}

func (f *funcs) injectNodeSelector(nodeSelector map[string]string, orig, mutable *corev1.PodSpec) {
	if f == nil || f.nodeSelectorInjector == nil || nodeSelector == nil || orig == nil || mutable == nil {
		return
	}

	f.nodeSelectorInjector.injectNodeSelector(nodeSelector, orig, mutable)
}

func (f *funcs) injectSchedulerName(schedulerName string, orig, mutable *corev1.PodSpec) {
	if f == nil || f.schedulerNameInjector == nil || schedulerName == "" || orig == nil || mutable == nil {
		return
	}

	f.schedulerNameInjector.injectSchedulerName(schedulerName, orig, mutable)
}

func (f *funcs) injectTolerations(tolerations []corev1.Toleration, orig, mutable *corev1.PodSpec) {
	if f == nil || f.tolerationsInjector == nil || tolerations == nil || orig == nil || mutable == nil {
		return
	}

	f.tolerationsInjector.injectTolerations(tolerations, orig, mutable)
}

func (f *funcs) injectPodSchedulingPolicyConfiguration(spc common.PodSchedulingPolicyConfiguration, orig, mutable *corev1.PodSpec) {
	if f == nil || spc == nil {
		return
	}

	f.injectAffinity(spc.GetAffinity(), orig, mutable)
	f.injectNodeName(spc.GetNodeName(), orig, mutable)
	f.injectNodeSelector(spc.GetNodeSelector(), orig, mutable)
	f.injectSchedulerName(spc.GetSchedulerName(), orig, mutable)
	f.injectTolerations(spc.GetTolerations(), orig, mutable)
}

func newDefaultPodSchedulingPolicyInjector() *funcs {
	return &funcs{
		affinityInjector:      affinityInjectorFunc(defaultInjectAffinity),
		nodeNameInjector:      nodeNameInjectorFunc(defaultInjectNodeName),
		nodeSelectorInjector:  nodeSelectorInjectorFunc(defaultInjectNodeSelector),
		schedulerNameInjector: schedulerNameInjectorFunc(defaultInjectSchedulerName),
		tolerationsInjector:   tolerationsInjectorFunc(defaultInjectTolerations),
	}
}

func defaultInjectAffinity(a *corev1.Affinity, orig, mutable *corev1.PodSpec) {
	if orig == nil || a == nil || mutable == nil {
		return
	}

	if mutable.Affinity == nil {
		mutable.Affinity = a.DeepCopy()
		return
	}

	defaultMergeNodeAffinity(a.NodeAffinity, mutable.Affinity)
	defaultInjectPodAffinity(a.PodAffinity, mutable.Affinity)
	defaultInjectPodAntiAffinity(a.PodAntiAffinity, mutable.Affinity)
}

func defaultMergeNodeAffinity(s *corev1.NodeAffinity, mutable *corev1.Affinity) {
	if s == nil {
		return
	}

	s = s.DeepCopy()
	if mutable.NodeAffinity == nil {
		mutable.NodeAffinity = s
		return
	}

	t := mutable.NodeAffinity

	mergeUniquePreferredSchedulingTerms(s.PreferredDuringSchedulingIgnoredDuringExecution, &(t.PreferredDuringSchedulingIgnoredDuringExecution))

	if t.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		t.RequiredDuringSchedulingIgnoredDuringExecution = s.RequiredDuringSchedulingIgnoredDuringExecution
	} else if s.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		mergeUniqueNodeSelectorTerms(s.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, &(t.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms))
	}
}

func mergeUniquePreferredSchedulingTerms(s []corev1.PreferredSchedulingTerm, mutable *[]corev1.PreferredSchedulingTerm) {
	var exists = func(p *corev1.PreferredSchedulingTerm) bool {
		var slice = *mutable
		for i := range slice {
			var ip = &slice[i]
			if reflect.DeepEqual(p, ip) {
				return true
			}
		}
		return false
	}

	for i := range s {
		var ip = &s[i]
		if !exists(ip) {
			*mutable = append(*mutable, *ip)
		}
	}
}

func mergeUniqueNodeSelectorTerms(s []corev1.NodeSelectorTerm, mutable *[]corev1.NodeSelectorTerm) {
	var exists = func(n *corev1.NodeSelectorTerm) bool {
		var slice = *mutable
		for i := range slice {
			var in = &slice[i]
			if reflect.DeepEqual(n, in) {
				return true
			}
		}
		return false
	}

	for i := range s {
		var in = &s[i]
		if !exists(in) {
			*mutable = append(*mutable, *in)
		}
	}
}

func defaultInjectPodAffinity(s *corev1.PodAffinity, mutable *corev1.Affinity) {
	if s == nil {
		return
	}

	s = s.DeepCopy()
	if mutable.PodAffinity == nil {
		mutable.PodAffinity = s
		return
	}

	t := mutable.PodAffinity

	mergeUniqueWeightedPodAffinityTerms(s.PreferredDuringSchedulingIgnoredDuringExecution, &(t.PreferredDuringSchedulingIgnoredDuringExecution))
	mergeUniquePodAffinityTerms(s.RequiredDuringSchedulingIgnoredDuringExecution, &(t.RequiredDuringSchedulingIgnoredDuringExecution))
}

func mergeUniqueWeightedPodAffinityTerms(s []corev1.WeightedPodAffinityTerm, mutable *[]corev1.WeightedPodAffinityTerm) {
	var exists = func(w *corev1.WeightedPodAffinityTerm) bool {
		var slice = *mutable
		for i := range slice {
			var iw = &slice[i]
			if reflect.DeepEqual(w, iw) {
				return true
			}
		}
		return false
	}

	for i := range s {
		var iw = &s[i]
		if !exists(iw) {
			*mutable = append(*mutable, *iw)
		}
	}
}

func mergeUniquePodAffinityTerms(s []corev1.PodAffinityTerm, mutable *[]corev1.PodAffinityTerm) {
	var exists = func(p *corev1.PodAffinityTerm) bool {
		var slice = *mutable
		for i := range slice {
			var ip = &slice[i]
			if reflect.DeepEqual(p, ip) {
				return true
			}
		}
		return false
	}

	for i := range s {
		var ip = &s[i]
		if !exists(ip) {
			*mutable = append(*mutable, *ip)
		}
	}
}

func defaultInjectPodAntiAffinity(s *corev1.PodAntiAffinity, mutable *corev1.Affinity) {
	if s == nil {
		return
	}

	s = s.DeepCopy()
	if mutable.PodAntiAffinity == nil {
		mutable.PodAntiAffinity = s
		return
	}

	t := mutable.PodAntiAffinity

	mergeUniqueWeightedPodAffinityTerms(s.PreferredDuringSchedulingIgnoredDuringExecution, &(t.PreferredDuringSchedulingIgnoredDuringExecution))
	mergeUniquePodAffinityTerms(s.RequiredDuringSchedulingIgnoredDuringExecution, &(t.RequiredDuringSchedulingIgnoredDuringExecution))
}

func defaultInjectNodeName(nodeName string, orig, mutable *corev1.PodSpec) {
	if orig == nil || nodeName == "" || mutable == nil || mutable.NodeName != "" {
		return
	}

	mutable.NodeName = nodeName
}

func defaultInjectNodeSelector(nodeSelector map[string]string, orig, mutable *corev1.PodSpec) {
	if orig == nil || nodeSelector == nil || mutable == nil {
		return
	}

	if mutable.NodeSelector == nil {
		mutable.NodeSelector = make(map[string]string, len(nodeSelector))
	}
	for k, v := range nodeSelector {
		if _, ok := mutable.NodeSelector[k]; ok {
			continue
		}
		mutable.NodeSelector[k] = v
	}
}

func defaultInjectSchedulerName(schedulerName string, orig, mutable *corev1.PodSpec) {
	if orig == nil || schedulerName == "" || mutable == nil || mutable.SchedulerName != "" {
		return
	}

	mutable.SchedulerName = schedulerName
}

func defaultInjectTolerations(tolerations []corev1.Toleration, orig, mutable *corev1.PodSpec) {
	if orig == nil || tolerations == nil || mutable == nil {
		return
	}

	var exists = func(t *corev1.Toleration) bool {
		for i := range mutable.Tolerations {
			var it = &mutable.Tolerations[i]
			if reflect.DeepEqual(t, it) {
				return true
			}
		}
		return false
	}

	for i := range tolerations {
		var it = &tolerations[i]
		if !exists(it) {
			mutable.Tolerations = append(mutable.Tolerations, *it)
		}
	}
}
