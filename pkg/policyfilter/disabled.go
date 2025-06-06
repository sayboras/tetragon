// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

// nolint:revive // prevent unused-parameter alert, disabled method obviously don't use args
package policyfilter

import (
	"errors"

	slimv1 "github.com/cilium/tetragon/pkg/k8s/slim/k8s/apis/meta/v1"
	"github.com/cilium/tetragon/pkg/labels"
	"github.com/cilium/tetragon/pkg/podhelpers"
	"k8s.io/client-go/tools/cache"
)

func DisabledState() State {
	return &disabled{}
}

type disabled struct {
}

func (s *disabled) AddPolicy(polID PolicyID, namespace string, podSelector *slimv1.LabelSelector,
	containerSelector *slimv1.LabelSelector) error {
	return errors.New("policyfilter is disabled")
}

func (s *disabled) DelPolicy(polID PolicyID) error {
	if polID == NoFilterPolicyID {
		return nil
	}
	return errors.New("policyfilter is disabled")
}

func (s *disabled) AddPodContainer(podID PodID, namespace, workload, kind string, podLabels labels.Labels,
	containerID string, cgID CgroupID, containerInfo podhelpers.ContainerInfo) error {
	return nil
}

func (s *disabled) UpdatePod(podID PodID, namespace, workload, kind string, podLabels labels.Labels,
	containerIDs []string, containerInfo []podhelpers.ContainerInfo) error {
	return nil
}

func (s *disabled) DelPodContainer(podID PodID, containerID string) error {
	return nil
}

func (s *disabled) DelPod(podID PodID) error {
	return nil
}

func (s *disabled) RegisterPodHandlers(podInformer cache.SharedIndexInformer) {
}

func (s *disabled) Close() error {
	return nil
}

func (s *disabled) GetNsId(stateID StateID) (*NSID, bool) {
	return nil, false
}

func (s *disabled) GetIdNs(id NSID) (StateID, bool) {
	return StateID(0), false
}
