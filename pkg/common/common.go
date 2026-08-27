// SPDX-FileCopyrightText: Contributors to the Gardener project
//
// SPDX-License-Identifier: Apache-2.0

package common

import (
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

// PodSchedulingPolicyConfiguration describes the interface for scheduling policy configuration.
type PodSchedulingPolicyConfiguration interface {
	runtime.Object
	GetNodeSelector() map[string]string
	GetNodeName() string
	GetAffinity() *corev1.Affinity
	GetSchedulerName() string
	GetTolerations() []corev1.Toleration
}
