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
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// podProcessorFactory implements the processFactory interface to mutate pod spec in pods.
type podProcessorFactory struct{}

func (pf *podProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   corev1.SchemeGroupVersion.Group,
		Version: corev1.SchemeGroupVersion.Version,
		Kind:    "Pod",
	}
}

func (pf *podProcessorFactory) newProcessor() processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &podCallbacks{},
	}
}

func (pf *podProcessorFactory) isMutating() bool {
	return true
}

// podCallbacks implements podSpecCallbacks
type podCallbacks struct {
	pod corev1.Pod
}

func (p *podCallbacks) getObject() runtime.Object {
	return &(p.pod)
}

func (p *podCallbacks) getPodSpec() *corev1.PodSpec {
	return &(p.pod.Spec)
}

func (p *podCallbacks) getPodLabels() map[string]string {
	return p.pod.Labels
}
