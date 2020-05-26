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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ClusterPodSchedulingPolicySpec defines the desired state of ClusterPodSchedulingPolicy
type ClusterPodSchedulingPolicySpec struct {
	// PodSelector selects the pods to which the ClusterPodSchedulingPolicy applies.
	// Any given pod might match the PodSelector of multiple podschedulingpolicies.
	// In such a case the matching policies will be merged. TODO explain how the merge happens.
	// If present empty, it selects all pods in all the namespaces selected by NamespaceSelector.
	// If not present it does not select any pods.
	//
	// +optional
	PodSelector *metav1.LabelSelector `json:"podSelector,omitempty" protobuf:"bytes,2,opt,name=podSelector"`
	// Selects Namespaces using cluster-scoped labels. This field follows standard label
	// selector semantics.
	// If present but empty, it selects all namespaces.
	// If not present, it does not select any namespaces and hence does not select any pods either.
	//
	// +optional
	NamespaceSelector *metav1.LabelSelector `json:"namespaceSelector,omitempty" protobuf:"bytes,2,opt,name=namespaceSelector"`
	// NodeSelector is a selector which must be true for the pod to fit on a node.
	// Selector which must match a node's labels for the pod to be scheduled on that node.
	// More info: https://kubernetes.io/docs/concepts/configuration/assign-pod-node/
	// +optional
	NodeSelector map[string]string `json:"nodeSelector,omitempty" protobuf:"bytes,7,rep,name=nodeSelector"`
	// NodeName is a request to schedule this pod onto a specific node. If it is non-empty,
	// the scheduler simply schedules this pod onto that node, assuming that it fits resource
	// requirements.
	// +optional
	NodeName string `json:"nodeName,omitempty" protobuf:"bytes,10,opt,name=nodeName"`
	// If specified, the pod's scheduling constraints
	// +optional
	Affinity *corev1.Affinity `json:"affinity,omitempty" protobuf:"bytes,18,opt,name=affinity"`
	// If specified, the pod will be dispatched by specified scheduler.
	// If not specified, the pod will be dispatched by default scheduler.
	// +optional
	SchedulerName string `json:"schedulerName,omitempty" protobuf:"bytes,19,opt,name=schedulerName"`
	// If specified, the pod's tolerations.
	// +optional
	Tolerations []corev1.Toleration `json:"tolerations,omitempty" protobuf:"bytes,22,opt,name=tolerations"`
}

// +kubebuilder:object:root=true
// +kubebuilder:resource:scope=Cluster

// ClusterPodSchedulingPolicy is the Schema for the clusterpodschedulingpolicies API
type ClusterPodSchedulingPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec ClusterPodSchedulingPolicySpec `json:"spec,omitempty"`
}

// +kubebuilder:object:root=true

// ClusterPodSchedulingPolicyList contains a list of ClusterPodSchedulingPolicy
type ClusterPodSchedulingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []ClusterPodSchedulingPolicy `json:"items"`
}

func init() {
	SchemeBuilder.Register(&ClusterPodSchedulingPolicy{}, &ClusterPodSchedulingPolicyList{})
}

// GetNodeSelector returns the configured NodeSelector.
func (s *ClusterPodSchedulingPolicy) GetNodeSelector() map[string]string {
	return s.Spec.NodeSelector
}

// GetNodeName returns the configured NodeName.
func (s *ClusterPodSchedulingPolicy) GetNodeName() string {
	return s.Spec.NodeName
}

// GetAffinity returns the configured Affinity.
func (s *ClusterPodSchedulingPolicy) GetAffinity() *corev1.Affinity {
	return s.Spec.Affinity
}

// GetSchedulerName returns the configured SchedulerName.
func (s *ClusterPodSchedulingPolicy) GetSchedulerName() string {
	return s.Spec.SchedulerName
}

// GetTolerations returns the configured Tolerations.
func (s *ClusterPodSchedulingPolicy) GetTolerations() []corev1.Toleration {
	return s.Spec.Tolerations
}
