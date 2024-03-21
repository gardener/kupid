// SPDX-FileCopyrightText: 2024 SAP SE or an SAP affiliate company and Gardener contributors
//
// SPDX-License-Identifier: Apache-2.0

package webhook

import (
	"github.com/go-logr/logr"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var testLog = logr.Discard()

// describeDefaultInjection describes the common pattern that the tests are described for default injection functions.
func describeDefaultInjection(
	text string,
	noValue interface{},
	newValue interface{},
	podSpecWithoutValues, podSpecWithConflictingOldValues, podSpecWithNonConflictingOldValues func() *corev1.PodSpec,
	inject func(newValue interface{}, orig, mutable *corev1.PodSpec),
	resetValueAfterInjection func(mutable, mutableBefore *corev1.PodSpec),
	mutateValues func(newValue interface{}, mutableBefore *corev1.PodSpec) *corev1.PodSpec) bool {

	var _ = Describe(text, func() {
		(func() {
			var entries = []TableEntry{
				Entry("all nil", noValue, nil, nil),
				Entry("no new value", noValue, podSpecWithoutValues(), nil),
				Entry("no new value", noValue, nil, podSpecWithoutValues()),
				Entry("no orig", newValue, nil, nil),
				Entry("no orig", newValue, nil, podSpecWithoutValues()),
			}
			if podSpecWithConflictingOldValues != nil {
				entries = append(entries, Entry("no new value", noValue, podSpecWithConflictingOldValues(), nil))
				entries = append(entries, Entry("no new value", noValue, podSpecWithConflictingOldValues(), podSpecWithConflictingOldValues()))
				entries = append(entries, Entry("no orig override", newValue, podSpecWithConflictingOldValues(), podSpecWithConflictingOldValues()))
			}

			DescribeTable("should not inject anything", func(newValue interface{}, orig, mutable *corev1.PodSpec) {
				var (
					origBefore, mutableBefore *corev1.PodSpec
				)
				if orig != nil {
					origBefore = orig.DeepCopy()
				}
				if mutable != nil {
					mutableBefore = mutable.DeepCopy()
				}

				Expect(inject).ToNot(BeNil())

				inject(newValue, orig, mutable)

				// orig should not be modified in any way.
				Expect(orig).To(Equal(origBefore))

				// mutable should not be modified in any way.
				Expect(mutable).To(Equal(mutableBefore))
			},
				entries...,
			)
		})()

		(func() {
			var entries = []TableEntry{
				Entry("no old values in mutable", newValue, podSpecWithoutValues(), podSpecWithoutValues()),
			}

			if podSpecWithNonConflictingOldValues != nil {
				entries = append(entries, Entry("with old values in mutable ", newValue, podSpecWithNonConflictingOldValues(), podSpecWithNonConflictingOldValues()))
			}

			DescribeTable("should inject new value", func(newValue interface{}, orig, mutable *corev1.PodSpec) {
				var (
					origBefore, mutableBefore *corev1.PodSpec
				)
				if orig != nil {
					origBefore = orig.DeepCopy()
				}
				if mutable != nil {
					mutableBefore = mutable.DeepCopy()
				}

				Expect(inject).ToNot(BeNil())
				Expect(resetValueAfterInjection).ToNot(BeNil())
				Expect(mutateValues).ToNot(BeNil())

				// there should be something to inject
				Expect(newValue).To(Or(Not(BeNil()), Not(BeEmpty())))

				inject(newValue, orig, mutable)

				// mutable should not be modified in any way.
				Expect(orig).To(Equal(origBefore))

				// except for the intended injected values nothing else should not be modified in any way.
				Expect((func() *corev1.PodSpec {
					m := mutable.DeepCopy()
					resetValueAfterInjection(m, mutableBefore.DeepCopy())
					return m
				})()).To(Equal(mutableBefore))

				Expect(mutable).To(Equal((func() *corev1.PodSpec {
					return mutateValues(newValue, mutableBefore.DeepCopy())
				})()))

				// repeat the injection
				mutableBefore = mutable.DeepCopy()

				inject(newValue, orig, mutable)

				Expect(mutable).To(Equal(mutableBefore))
			},
				entries...,
			)
		})()
	})
	return true
}

var _ = Describe("defaultInjectAffinity", func() {
	(func() {
		var podSpec = &corev1.PodSpec{
			Affinity: &corev1.Affinity{
				NodeAffinity: &corev1.NodeAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
						{
							Preference: corev1.NodeSelectorTerm{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "newPreferredExpressionKey",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"newPreferredExpressionValue"},
									},
								},
								MatchFields: []corev1.NodeSelectorRequirement{
									{
										Key:      "newPreferredFieldKey",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"newPreferredFieldValue"},
									},
								},
							},
							Weight: 100,
						},
					},
					RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
						NodeSelectorTerms: []corev1.NodeSelectorTerm{
							{
								MatchExpressions: []corev1.NodeSelectorRequirement{
									{
										Key:      "newRequiredExpressionKey",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"newRequiredExpressionValue"},
									},
								},
								MatchFields: []corev1.NodeSelectorRequirement{
									{
										Key:      "newRequiredFieldKey",
										Operator: corev1.NodeSelectorOpIn,
										Values:   []string{"newRequiredFieldValue"},
									},
								},
							},
						},
					},
				},
			},
		}

		var _ = describeDefaultInjection(
			"node affinity",
			(func() *corev1.Affinity { return nil })(),
			podSpec.Affinity,
			func() *corev1.PodSpec { return &corev1.PodSpec{} },
			func() *corev1.PodSpec { return podSpec.DeepCopy() },
			func() *corev1.PodSpec {
				return &corev1.PodSpec{
					Affinity: &corev1.Affinity{
						NodeAffinity: &corev1.NodeAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.PreferredSchedulingTerm{
								{
									Preference: corev1.NodeSelectorTerm{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "oldPreferredExpressionKey",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"oldPreferredExpressionValue"},
											},
										},
										MatchFields: []corev1.NodeSelectorRequirement{
											{
												Key:      "oldPreferredFieldKey",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"oldPreferredFieldValue"},
											},
										},
									},
									Weight: 100,
								},
							},
							RequiredDuringSchedulingIgnoredDuringExecution: &corev1.NodeSelector{
								NodeSelectorTerms: []corev1.NodeSelectorTerm{
									{
										MatchExpressions: []corev1.NodeSelectorRequirement{
											{
												Key:      "oldRequiredExpressionKey",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"oldRequiredExpressionValue"},
											},
										},
										MatchFields: []corev1.NodeSelectorRequirement{
											{
												Key:      "oldRequiredFieldKey",
												Operator: corev1.NodeSelectorOpIn,
												Values:   []string{"oldRequiredFieldValue"},
											},
										},
									},
								},
							},
						},
					},
				}
			},
			func(newValue interface{}, orig, mutable *corev1.PodSpec) {
				Expect(newValue).To(BeAssignableToTypeOf(&corev1.Affinity{}))
				defaultInjectAffinity(newValue.(*corev1.Affinity), orig, mutable, testLog)
			},
			func(mutable, mutableBefore *corev1.PodSpec) {
				mutable.Affinity = mutableBefore.Affinity
			},
			func(newValue interface{}, mutableBefore *corev1.PodSpec) *corev1.PodSpec {
				Expect(newValue).To(BeAssignableToTypeOf(&corev1.Affinity{}))

				t := newValue.(*corev1.Affinity)
				Expect(t).ToNot(BeNil())
				Expect(t.NodeAffinity).ToNot(BeNil())

				if mutableBefore.Affinity == nil {
					mutableBefore.Affinity = t
				} else if mutableBefore.Affinity.NodeAffinity == nil {
					mutableBefore.Affinity.NodeAffinity = t.NodeAffinity
				} else {
					mutableBefore.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(mutableBefore.Affinity.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution, t.NodeAffinity.PreferredDuringSchedulingIgnoredDuringExecution...)

					if mutableBefore.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution == nil {
						mutableBefore.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution = t.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution
					} else if t.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution != nil {
						mergedNSTs := make([]corev1.NodeSelectorTerm, 0)
						for _, policyNST := range t.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
							for _, podNST := range mutableBefore.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms {
								policyNST.MatchExpressions = append(policyNST.MatchExpressions, podNST.MatchExpressions...)
								policyNST.MatchFields = append(policyNST.MatchFields, podNST.MatchFields...)
							}
							mergedNSTs = append(mergedNSTs, policyNST)
						}
						mutableBefore.Affinity.NodeAffinity.RequiredDuringSchedulingIgnoredDuringExecution.NodeSelectorTerms = mergedNSTs
					}
				}
				return mutableBefore
			},
		)
	})()

	(func() {
		var podSpec = &corev1.PodSpec{
			Affinity: &corev1.Affinity{
				PodAffinity: &corev1.PodAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"preferredNewKey": "preferredNewValue",
									},
								},
								Namespaces:  []string{"preferredNewNamespace"},
								TopologyKey: "preferredNewTopologyKey",
							},
							Weight: 100,
						},
					},
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"requiredNewKey": "requiredNewValue",
								},
							},
							Namespaces:  []string{"requiredNewNamespace"},
							TopologyKey: "requiredNewTopologyKey",
						},
					},
				},
			},
		}

		var _ = describeDefaultInjection(
			"pod affinity",
			(func() *corev1.Affinity { return nil })(),
			podSpec.Affinity,
			func() *corev1.PodSpec { return &corev1.PodSpec{} },
			func() *corev1.PodSpec { return podSpec.DeepCopy() },
			func() *corev1.PodSpec {
				return &corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAffinity: &corev1.PodAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									PodAffinityTerm: corev1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{
												"preferredOldKey": "preferredOldValue",
											},
										},
										Namespaces:  []string{"preferredOldNamespace"},
										TopologyKey: "preferredOldTopologyKey",
									},
									Weight: 10,
								},
							},
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"requiredOldKey": "requiredOldValue",
										},
									},
									Namespaces:  []string{"requiredOldNamespace"},
									TopologyKey: "requiredOldTopologyKey",
								},
							},
						},
					},
				}
			},
			func(newValue interface{}, orig, mutable *corev1.PodSpec) {
				Expect(newValue).To(BeAssignableToTypeOf(&corev1.Affinity{}))
				defaultInjectAffinity(newValue.(*corev1.Affinity), orig, mutable, testLog)
			},
			func(mutable, mutableBefore *corev1.PodSpec) {
				mutable.Affinity = mutableBefore.Affinity
			},
			func(newValue interface{}, mutableBefore *corev1.PodSpec) *corev1.PodSpec {
				Expect(newValue).To(BeAssignableToTypeOf(&corev1.Affinity{}))

				t := newValue.(*corev1.Affinity)
				Expect(t).ToNot(BeNil())
				Expect(t.PodAffinity).ToNot(BeNil())

				if mutableBefore.Affinity == nil {
					mutableBefore.Affinity = t
				} else if mutableBefore.Affinity.PodAffinity == nil {
					mutableBefore.Affinity.PodAffinity = t.PodAffinity
				} else {
					mutableBefore.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(mutableBefore.Affinity.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution, t.PodAffinity.PreferredDuringSchedulingIgnoredDuringExecution...)
					mutableBefore.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(mutableBefore.Affinity.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution, t.PodAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
				}
				return mutableBefore
			},
		)
	})()

	(func() {
		var podSpec = &corev1.PodSpec{
			Affinity: &corev1.Affinity{
				PodAntiAffinity: &corev1.PodAntiAffinity{
					PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
						{
							PodAffinityTerm: corev1.PodAffinityTerm{
								LabelSelector: &metav1.LabelSelector{
									MatchLabels: map[string]string{
										"preferredNewKey": "preferredNewValue",
									},
								},
								Namespaces:  []string{"preferredNewNamespace"},
								TopologyKey: "preferredNewTopologyKey",
							},
							Weight: 100,
						},
					},
					RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
						{
							LabelSelector: &metav1.LabelSelector{
								MatchLabels: map[string]string{
									"requiredNewKey": "requiredNewValue",
								},
							},
							Namespaces:  []string{"requiredNewNamespace"},
							TopologyKey: "requiredNewTopologyKey",
						},
					},
				},
			},
		}

		var _ = describeDefaultInjection(
			"pod anti-affinity",
			(func() *corev1.Affinity { return nil })(),
			podSpec.Affinity,
			func() *corev1.PodSpec { return &corev1.PodSpec{} },
			func() *corev1.PodSpec { return podSpec.DeepCopy() },
			func() *corev1.PodSpec {
				return &corev1.PodSpec{
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{
								{
									PodAffinityTerm: corev1.PodAffinityTerm{
										LabelSelector: &metav1.LabelSelector{
											MatchLabels: map[string]string{
												"preferredOldKey": "preferredOldValue",
											},
										},
										Namespaces:  []string{"preferredOldNamespace"},
										TopologyKey: "preferredOldTopologyKey",
									},
									Weight: 10,
								},
							},
							RequiredDuringSchedulingIgnoredDuringExecution: []corev1.PodAffinityTerm{
								{
									LabelSelector: &metav1.LabelSelector{
										MatchLabels: map[string]string{
											"requiredOldKey": "requiredOldValue",
										},
									},
									Namespaces:  []string{"requiredOldNamespace"},
									TopologyKey: "requiredOldTopologyKey",
								},
							},
						},
					},
				}
			},
			func(newValue interface{}, orig, mutable *corev1.PodSpec) {
				Expect(newValue).To(BeAssignableToTypeOf(&corev1.Affinity{}))
				defaultInjectAffinity(newValue.(*corev1.Affinity), orig, mutable, testLog)
			},
			func(mutable, mutableBefore *corev1.PodSpec) {
				mutable.Affinity = mutableBefore.Affinity
			},
			func(newValue interface{}, mutableBefore *corev1.PodSpec) *corev1.PodSpec {
				Expect(newValue).To(BeAssignableToTypeOf(&corev1.Affinity{}))

				t := newValue.(*corev1.Affinity)
				Expect(t).ToNot(BeNil())
				Expect(t.PodAntiAffinity).ToNot(BeNil())

				if mutableBefore.Affinity == nil {
					mutableBefore.Affinity = t
				} else if mutableBefore.Affinity.PodAntiAffinity == nil {
					mutableBefore.Affinity.PodAntiAffinity = t.PodAntiAffinity
				} else {
					mutableBefore.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution = append(mutableBefore.Affinity.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution, t.PodAntiAffinity.PreferredDuringSchedulingIgnoredDuringExecution...)
					mutableBefore.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution = append(mutableBefore.Affinity.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution, t.PodAntiAffinity.RequiredDuringSchedulingIgnoredDuringExecution...)
				}
				return mutableBefore
			},
		)
	})()
})

var _ = describeDefaultInjection(
	"defaultInjectNodeName",
	"",
	"newNodeName",
	func() *corev1.PodSpec { return &corev1.PodSpec{} },
	func() *corev1.PodSpec {
		return &corev1.PodSpec{
			NodeName: "oldNodeName",
		}
	},
	nil,
	func(newValue interface{}, orig, mutable *corev1.PodSpec) {
		Expect(newValue).To(BeAssignableToTypeOf(""))
		defaultInjectNodeName(newValue.(string), orig, mutable)
	},
	func(mutable, mutableBefore *corev1.PodSpec) {
		mutable.NodeName = mutableBefore.NodeName
	},
	func(newValue interface{}, mutableBefore *corev1.PodSpec) *corev1.PodSpec {
		Expect(newValue).To(BeAssignableToTypeOf(""))

		mutableBefore.NodeName = newValue.(string)
		return mutableBefore
	},
)

var _ = (func() bool {
	var podSpec = &corev1.PodSpec{
		NodeSelector: map[string]string{
			"newKey": "newValue",
		},
	}
	var _ = describeDefaultInjection(
		"defaultInjectNodeSelector",
		(func() map[string]string { return nil })(),
		podSpec.NodeSelector,
		func() *corev1.PodSpec { return &corev1.PodSpec{} },
		func() *corev1.PodSpec { return podSpec.DeepCopy() },
		func() *corev1.PodSpec {
			return &corev1.PodSpec{
				NodeSelector: map[string]string{
					"oldKey": "oldValue",
				},
			}
		},
		func(newValue interface{}, orig, mutable *corev1.PodSpec) {
			Expect(newValue).To(BeAssignableToTypeOf(map[string]string{}))
			defaultInjectNodeSelector(newValue.(map[string]string), orig, mutable)
		},
		func(mutable, mutableBefore *corev1.PodSpec) {
			mutable.NodeSelector = mutableBefore.NodeSelector
		},
		func(newValue interface{}, mutableBefore *corev1.PodSpec) *corev1.PodSpec {
			Expect(newValue).To(BeAssignableToTypeOf(map[string]string{}))

			nvm := newValue.(map[string]string)
			if len(mutableBefore.NodeSelector) > 0 {
				for k, v := range nvm {
					mutableBefore.NodeSelector[k] = v
				}
			} else {
				mutableBefore.NodeSelector = nvm
			}
			return mutableBefore
		},
	)
	return false
})()

var _ = describeDefaultInjection(
	"defaultInjectSchedulerName",
	"",
	"newSchedulerName",
	func() *corev1.PodSpec { return &corev1.PodSpec{} },
	func() *corev1.PodSpec {
		return &corev1.PodSpec{
			SchedulerName: "oldScheduler",
		}
	},
	nil,
	func(newValue interface{}, orig, mutable *corev1.PodSpec) {
		Expect(newValue).To(BeAssignableToTypeOf(""))
		defaultInjectSchedulerName(newValue.(string), orig, mutable)
	},
	func(mutable, mutableBefore *corev1.PodSpec) {
		mutable.SchedulerName = mutableBefore.SchedulerName
	},
	func(newValue interface{}, mutableBefore *corev1.PodSpec) *corev1.PodSpec {
		Expect(newValue).To(BeAssignableToTypeOf(""))

		mutableBefore.SchedulerName = newValue.(string)
		return mutableBefore
	},
)

var _ = (func() bool {
	var podSpec = &corev1.PodSpec{
		Tolerations: []corev1.Toleration{
			{
				Key:      "newKey",
				Operator: corev1.TolerationOpEqual,
				Value:    "newValue",
				Effect:   corev1.TaintEffectPreferNoSchedule,
			},
		},
	}
	var _ = describeDefaultInjection(
		"defaultInjectTolerations",
		(func() []corev1.Toleration { return nil })(),
		podSpec.Tolerations,
		func() *corev1.PodSpec { return &corev1.PodSpec{} },
		func() *corev1.PodSpec { return podSpec.DeepCopy() },
		func() *corev1.PodSpec {
			return &corev1.PodSpec{
				Tolerations: []corev1.Toleration{
					{
						Key:      "oldKey",
						Operator: corev1.TolerationOpEqual,
						Value:    "oldValue",
						Effect:   corev1.TaintEffectPreferNoSchedule,
					},
				},
			}
		},
		func(newValue interface{}, orig, mutable *corev1.PodSpec) {
			Expect(newValue).To(BeAssignableToTypeOf([]corev1.Toleration{}))
			defaultInjectTolerations(newValue.([]corev1.Toleration), orig, mutable)
		},
		func(mutable, mutableBefore *corev1.PodSpec) {
			mutable.Tolerations = mutableBefore.Tolerations
		},
		func(newValue interface{}, mutableBefore *corev1.PodSpec) *corev1.PodSpec {
			Expect(newValue).To(BeAssignableToTypeOf([]corev1.Toleration{}))

			t := newValue.([]corev1.Toleration)
			mutableBefore.Tolerations = append(mutableBefore.Tolerations, t...)
			return mutableBefore
		},
	)
	return false
})()

var _ = Describe("Cartesian product of nodeSelectorTerms between pod spec and policy", func() {
	var (
		podNSTs          []corev1.NodeSelectorTerm
		policyNSTs       []corev1.NodeSelectorTerm
		expectedNSTSlice []corev1.NodeSelectorTerm

		MatchExpressionA = corev1.NodeSelectorRequirement{
			Key:      "a",
			Operator: "In",
			Values:   []string{"A"},
		}
		MatchExpressionB = corev1.NodeSelectorRequirement{
			Key:      "b",
			Operator: "In",
			Values:   []string{"B"},
		}
		MatchExpressionC = corev1.NodeSelectorRequirement{
			Key:      "c",
			Operator: "In",
			Values:   []string{"C"},
		}
		MatchExpressionD = corev1.NodeSelectorRequirement{
			Key:      "d",
			Operator: "In",
			Values:   []string{"D"},
		}
		MatchExpressionE = corev1.NodeSelectorRequirement{
			Key:      "e",
			Operator: "In",
			Values:   []string{"E"},
		}
		MatchFieldF = corev1.NodeSelectorRequirement{
			Key:      "metadata.name",
			Operator: "In",
			Values:   []string{"F"},
		}
	)

	BeforeEach(func() {
		podNSTs = []corev1.NodeSelectorTerm{}
		policyNSTs = []corev1.NodeSelectorTerm{}
	})

	Context("podNSTs is empty, policyNSTs is empty", func() {
		It("should return empty NST slice", func() {
			expectedNSTSlice = []corev1.NodeSelectorTerm{}
			Expect(mergeUniqueNodeSelectorTerms(policyNSTs, podNSTs, testLog)).To(Equal(expectedNSTSlice))
		})
	})

	Context("podNSTs is empty, policyNSTs is non-empty", func() {
		It("should return policyNSTs", func() {
			Expect(mergeUniqueNodeSelectorTerms(policyNSTs, podNSTs, testLog)).To(Equal(policyNSTs))
		})
	})

	Context("podNSTs is non-empty, policyNSTs is empty", func() {
		It("should return podNSTs", func() {
			Expect(mergeUniqueNodeSelectorTerms(policyNSTs, podNSTs, testLog)).To(Equal(podNSTs))
		})
	})

	Context("podNST is non-empty, policyNST is non-empty", func() {
		BeforeEach(func() {
			podNSTs = []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						MatchExpressionA,
						MatchExpressionB,
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						MatchExpressionC,
					},
				},
			}
			policyNSTs = []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						MatchExpressionD,
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						MatchExpressionE,
					},
					MatchFields: []corev1.NodeSelectorRequirement{
						MatchFieldF,
					},
				},
			}
		})
		It("should return correct NST slice", func() {
			expectedNSTSlice = []corev1.NodeSelectorTerm{
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						MatchExpressionD,
						MatchExpressionA,
						MatchExpressionB,
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						MatchExpressionE,
						MatchExpressionA,
						MatchExpressionB,
					},
					MatchFields: []corev1.NodeSelectorRequirement{
						MatchFieldF,
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						MatchExpressionD,
						MatchExpressionC,
					},
				},
				{
					MatchExpressions: []corev1.NodeSelectorRequirement{
						MatchExpressionE,
						MatchExpressionC,
					},
					MatchFields: []corev1.NodeSelectorRequirement{
						MatchFieldF,
					},
				},
			}
			Expect(mergeUniqueNodeSelectorTerms(policyNSTs, podNSTs, testLog)).To(Equal(expectedNSTSlice))
		})
	})

	Context("podNST is empty, policyNST is empty", func() {
		It("should return empty NST slice", func() {
			expectedNSTSlice = []corev1.NodeSelectorTerm{}
			Expect(mergeUniqueNodeSelectorTerms(policyNSTs, podNSTs, testLog)).To(Equal(expectedNSTSlice))
		})
	})

	Context("podNST is empty, policyNST is empty", func() {
		It("should return empty NST slice", func() {
			expectedNSTSlice = []corev1.NodeSelectorTerm{}
			Expect(mergeUniqueNodeSelectorTerms(policyNSTs, podNSTs, testLog)).To(Equal(expectedNSTSlice))
		})
	})
})
