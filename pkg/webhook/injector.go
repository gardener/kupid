// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"reflect"

	"github.com/gardener/kupid/pkg/common"
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
)

type affinityInjector interface {
	injectAffinity(affinity *corev1.Affinity, orig, mutable *corev1.PodSpec, log logr.Logger)
}

type affinityInjectorFunc func(affinity *corev1.Affinity, orig, mutable *corev1.PodSpec, log logr.Logger)

func (f affinityInjectorFunc) injectAffinity(affinity *corev1.Affinity, orig, mutable *corev1.PodSpec, log logr.Logger) {
	if f == nil {
		return
	}

	f(affinity, orig, mutable, log)
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
	injectPodSchedulingPolicyConfiguration(spc common.PodSchedulingPolicyConfiguration, orig, mutable *corev1.PodSpec, log logr.Logger)
}

// funcs implements all the injectors.
type funcs struct {
	affinityInjector
	nodeNameInjector
	nodeSelectorInjector
	schedulerNameInjector
	tolerationsInjector
}

func (f *funcs) injectAffinity(affinity *corev1.Affinity, orig, mutable *corev1.PodSpec, log logr.Logger) {
	if f == nil || f.affinityInjector == nil || affinity == nil || orig == nil || mutable == nil {
		return
	}

	f.affinityInjector.injectAffinity(affinity, orig, mutable, log)
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

func (f *funcs) injectPodSchedulingPolicyConfiguration(spc common.PodSchedulingPolicyConfiguration, orig, mutable *corev1.PodSpec, log logr.Logger) {
	if f == nil || spc == nil {
		return
	}

	f.injectAffinity(spc.GetAffinity(), orig, mutable, log)
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

func defaultInjectAffinity(a *corev1.Affinity, orig, mutable *corev1.PodSpec, log logr.Logger) {
	if orig == nil || a == nil || mutable == nil {
		return
	}

	if mutable.Affinity == nil {
		mutable.Affinity = a.DeepCopy()
		return
	}

	defaultMergeNodeAffinity(a.NodeAffinity, mutable.Affinity, log)
	defaultInjectPodAffinity(a.PodAffinity, mutable.Affinity)
	defaultInjectPodAntiAffinity(a.PodAntiAffinity, mutable.Affinity)
}

func defaultMergeNodeAffinity(s *corev1.NodeAffinity, mutable *corev1.Affinity, log logr.Logger) {
	if s == nil {
		return
	}

	s = s.DeepCopy()
	if mutable.NodeAffinity == nil {
		mutable.NodeAffinity = s
		log.Info("(defaultMergeNodeAffinity) existing NodeAffinity is absent, initializing with affinity as defined in cluster/pod scheduling policy", "policyConfiguration", s)
		return
	}

	t := mutable.NodeAffinity

	mergeUniquePreferredSchedulingTerms(s.PreferredDuringSchedulingIgnoredDuringExecution, &(t.PreferredDuringSchedulingIgnoredDuringExecution), log)

	if t.RequiredDuringSchedulingIgnoredDuringExecution == nil {
		t.RequiredDuringSchedulingIgnoredDuringExecution = s.RequiredDuringSchedulingIgnoredDuringExecution
	} else if s.RequiredDuringSchedulingIgnoredDuringExecution != nil {
		t.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = mergeUniqueNodeSelectorTerms(s.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, t.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms, log)
	}
}

func mergeUniquePreferredSchedulingTerms(s []corev1.PreferredSchedulingTerm, mutable *[]corev1.PreferredSchedulingTerm, log logr.Logger) {
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
	delta := make([]corev1.PreferredSchedulingTerm, 0, len(s))
	for i := range s {
		var ip = &s[i]
		if !exists(ip) {
			*mutable = append(*mutable, *ip)
			delta = append(delta, *ip)
		}
	}
	if len(delta) > 0 {
		log.Info("Delta PreferredSchedulingTerms added", "PreferredSchedulingTerms", delta)
	} else {
		log.Info("No change identified for PreferredSchedulingTerms")
	}
}

func mergeUniqueNodeSelectorTerms(policyNSTs, podNSTs []corev1.NodeSelectorTerm, log logr.Logger) []corev1.NodeSelectorTerm {
	if podNSTs == nil || len(podNSTs) == 0 {
		return policyNSTs
	}
	if policyNSTs == nil || len(policyNSTs) == 0 {
		return podNSTs
	}

	mergedPodNSTs := make([]corev1.NodeSelectorTerm, 0, len(policyNSTs)*len(podNSTs))
	for _, podNST := range podNSTs {
		log.V(1).Info("creating cartesian product of NodeSelectorTerms from both Pod Spec and Policy Spec")
		mergedPodNSTs = append(mergedPodNSTs, createPodNSTCartesianProduct(podNST, policyNSTs, log)...)
	}
	printDeltaUniqueNodeSelectorTerms(podNSTs, mergedPodNSTs, log)
	return mergedPodNSTs
}

func printDeltaUniqueNodeSelectorTerms(podNSTs, updatedNSTs []corev1.NodeSelectorTerm, log logr.Logger) {
	var contains = func(podNSTs []corev1.NodeSelectorTerm, updatedNST corev1.NodeSelectorTerm) bool {
		for _, podNST := range podNSTs {
			if reflect.DeepEqual(podNST, updatedNST) {
				return true
			}
		}
		return false
	}

	delta := make([]corev1.NodeSelectorTerm, 0, len(updatedNSTs))
	for _, updatedNST := range updatedNSTs {
		if !contains(podNSTs, updatedNST) {
			delta = append(delta, updatedNST)
		}
	}
	if len(delta) > 0 {
		log.Info("Mutations made to RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms", "mutated NSTs", delta)
	} else {
		log.Info("No mutations done to RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms")
	}
}

// createPodNSTCartesianProduct returns a cartesian product of a single pod NST with all policy NSTs
func createPodNSTCartesianProduct(podNST corev1.NodeSelectorTerm, policyNSTs []corev1.NodeSelectorTerm, log logr.Logger) []corev1.NodeSelectorTerm {
	mergedNSTs := make([]corev1.NodeSelectorTerm, 0, len(policyNSTs))
	for _, policyNST := range policyNSTs {
		var mergedNST corev1.NodeSelectorTerm
		mergedNST.MatchExpressions = mergeNSRs(podNST.MatchExpressions, policyNST.MatchExpressions, log)
		mergedNST.MatchFields = mergeNSRs(podNST.MatchFields, policyNST.MatchFields, log)
		log.V(1).Info("(createPodNSTCartesianProduct) merged NSTs", "policyNST", policyNST, "merged-match-expressions", mergedNST.MatchExpressions, "merged-match-fields", mergedNST.MatchFields)
		mergedNSTs = append(mergedNSTs, mergedNST)
	}
	return mergedNSTs
}

func mergeNSRs(podNSRs, policyNSRs []corev1.NodeSelectorRequirement, log logr.Logger) []corev1.NodeSelectorRequirement {
	mergedNSRs := make([]corev1.NodeSelectorRequirement, len(policyNSRs))
	copy(mergedNSRs, policyNSRs)

	// contains checks if podNSR is contained in policyNSRs
	var contains = func(policyNSRs []corev1.NodeSelectorRequirement, podNSR corev1.NodeSelectorRequirement) bool {
		for _, policyNSR := range policyNSRs {
			if reflect.DeepEqual(policyNSR, podNSR) {
				return true
			}
		}
		return false
	}

	for _, podNSR := range podNSRs {
		if !contains(policyNSRs, podNSR) && !contains(mergedNSRs, podNSR) {
			log.V(1).Info("(mergedNSRs) podNSR under consideration is not contained in mergedNSRs", "podNSR", podNSR, "policyNSRs", policyNSRs, "mergedNSRs", mergedNSRs)
			mergedNSRs = append(mergedNSRs, podNSR)
		}
	}

	if len(mergedNSRs) == 0 {
		return nil
	}

	log.V(1).Info("(mergedNSRs) result", "mergedNSRs", mergedNSRs)
	return mergedNSRs
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
