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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// replicaSetProcessorFactory implements the processFactory interface to mutate pod template spec in replicasets.
type replicaSetProcessorFactory struct{}

func (pf *replicaSetProcessorFactory) kind() metav1.GroupVersionKind {
	return metav1.GroupVersionKind{
		Group:   appsv1.SchemeGroupVersion.Group,
		Version: appsv1.SchemeGroupVersion.Version,
		Kind:    "ReplicaSet",
	}
}

func (pf *replicaSetProcessorFactory) newProcessor(logger logr.Logger) processor {
	return &podSpecProcessorImpl{
		podSpecCallbacks: &replicaSetCallbacks{},
		logger:           logger,
	}
}

func (pf *replicaSetProcessorFactory) isMutating() bool {
	return true
}

// replicaSetCallbacks implements podSpecCallbacks
type replicaSetCallbacks struct {
	replicaSet appsv1.ReplicaSet
}

func (r *replicaSetCallbacks) getObject() runtime.Object {
	return &(r.replicaSet)
}

func (r *replicaSetCallbacks) getPodSpec() *corev1.PodSpec {
	return &(r.replicaSet.Spec.Template.Spec)
}

func (r *replicaSetCallbacks) getPodLabels() map[string]string {
	return r.replicaSet.Spec.Template.Labels
}
