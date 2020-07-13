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
	"github.com/go-logr/logr"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// replicationControllerProcessorFactory implements the processFactory interface to mutate pod template spec in replicationcontrollers.
type replicationControllerProcessorFactory struct{}

func (pf *replicationControllerProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   corev1.SchemeGroupVersion.Group,
		Version: corev1.SchemeGroupVersion.Version,
		Kind:    "ReplicationController",
	}
}

func (pf *replicationControllerProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &replicationControllerCallbacks{},
		logger:           logger,
	}
}

func (pf *replicationControllerProcessorFactory) isMutating() bool {
	return true
}

// replicationControllerCallbacks implements podSpecCallbacks
type replicationControllerCallbacks struct {
	replicationController corev1.ReplicationController
}

func (r *replicationControllerCallbacks) getObject() runtime.Object {
	return &(r.replicationController)
}

func (r *replicationControllerCallbacks) getPodSpec() *corev1.PodSpec {
	return &(r.replicationController.Spec.Template.Spec)
}

func (r *replicationControllerCallbacks) getPodLabels() map[string]string {
	return r.replicationController.Spec.Template.Labels
}
