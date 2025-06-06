// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

package v1alpha1

import (
	"fmt"

	ciliumio "github.com/cilium/tetragon/pkg/k8s/apis/cilium.io"
	slimv1 "github.com/cilium/tetragon/pkg/k8s/slim/k8s/apis/meta/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	// Tracing Policy (TP)

	// TPPluralName is the plural name of Cilium Tracing Policy
	TPPluralName = "tracingpolicies"

	// TPKindDefinition is the kind name of Cilium Tracing Policy
	TPKindDefinition = "TracingPolicy"

	// TPName is the full name of Cilium Egress NAT Policy
	TPName = TPPluralName + "." + ciliumio.GroupName

	// TPNamespacedPluralName is the plural name of Cilium Tracing Policy
	TPNamespacedPluralName = "tracingpoliciesnamespaced"

	// TPNamespacedName
	TPNamespacedName = TPNamespacedPluralName + "." + ciliumio.GroupName

	// TPKindDefinition is the kind name of Cilium Tracing Policy
	TPNamespacedKindDefinition = "TracingPolicyNamespaced"
)

// +genclient
// +genclient:noStatus
// +genclient:nonNamespaced
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories={tetragon},singular="tracingpolicy",path="tracingpolicies",scope="Cluster",shortName={tgtp}
type TracingPolicy struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	// Tracing policy specification.
	Spec TracingPolicySpec `json:"spec"`
}

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
// +kubebuilder:resource:categories={tetragon},singular="tracingpolicynamespaced",path="tracingpoliciesnamespaced",scope="Namespaced",shortName={tgtpn}
type TracingPolicyNamespaced struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata"`
	// Tracing policy specification.
	Spec TracingPolicySpec `json:"spec"`
}

func (tp *TracingPolicyNamespaced) TpSpec() *TracingPolicySpec {
	return &tp.Spec
}

func (tp *TracingPolicyNamespaced) TpInfo() string {
	return fmt.Sprintf("%s (object:%d/%s) (type:%s/%s)", tp.ObjectMeta.Name, tp.ObjectMeta.Generation, tp.ObjectMeta.UID, tp.TypeMeta.Kind, tp.TypeMeta.APIVersion)
}

func (tp *TracingPolicyNamespaced) TpName() string {
	return tp.ObjectMeta.Name
}

func (tp *TracingPolicyNamespaced) TpNamespace() string {
	return tp.ObjectMeta.Namespace
}

type TracingPolicySpec struct {
	// +kubebuilder:validation:Optional
	// A list of kprobe specs.
	KProbes []KProbeSpec `json:"kprobes,omitempty"`
	// +kubebuilder:validation:Optional
	// A list of tracepoint specs.
	Tracepoints []TracepointSpec `json:"tracepoints,omitempty"`
	// +kubebuilder:validation:Optional
	// Enable loader events
	Loader bool `json:"loader,omitempty"`
	// +kubebuilder:validation:Optional
	// A list of uprobe specs.
	UProbes []UProbeSpec `json:"uprobes,omitempty"`
	// +kubebuilder:validation:Optional
	// A list of uprobe specs.
	LsmHooks []LsmHookSpec `json:"lsmhooks,omitempty"`

	// +kubebuilder:validation:Optional
	// PodSelector selects pods that this policy applies to
	PodSelector *slimv1.LabelSelector `json:"podSelector,omitempty"`

	// +kubebuilder:validation:Optional
	// ContainerSelector selects containers that this policy applies to.
	// A map of container fields will be constructed in the same way as a map of labels.
	// The name of the field represents the label "key", and the value of the field - label "value".
	// Currently, only the "name" field is supported.
	ContainerSelector *slimv1.LabelSelector `json:"containerSelector,omitempty"`

	// +kubebuilder:validation:Optional
	// A list of list specs.
	Lists []ListSpec `json:"lists,omitempty"`

	// +kubebuilder:validation:Optional
	// A enforcer spec.
	Enforcers []EnforcerSpec `json:"enforcers,omitempty"`

	// +kubebuilder:validation:Optional
	// A list of overloaded options
	Options []OptionSpec `json:"options,omitempty"`
}

func (tp *TracingPolicy) TpName() string {
	return tp.ObjectMeta.Name
}

func (tp *TracingPolicy) TpSpec() *TracingPolicySpec {
	return &tp.Spec
}

func (tp *TracingPolicy) TpInfo() string {
	return fmt.Sprintf("%s (object:%d/%s) (type:%s/%s)", tp.ObjectMeta.Name, tp.ObjectMeta.Generation, tp.ObjectMeta.UID, tp.TypeMeta.Kind, tp.TypeMeta.APIVersion)
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TracingPolicyList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []TracingPolicy `json:"items"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object
type TracingPolicyNamespacedList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`
	Items           []TracingPolicyNamespaced `json:"items"`
}
