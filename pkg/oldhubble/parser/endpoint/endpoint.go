// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Hubble

package endpoint

import (
	"net"
	"sort"

	"github.com/cilium/cilium/api/v1/models"
	"github.com/cilium/cilium/pkg/hubble/k8s"
	"github.com/cilium/cilium/pkg/identity"
	monitorAPI "github.com/cilium/cilium/pkg/monitor/api"

	v1 "github.com/cilium/tetragon/pkg/oldhubble/api/v1"
)

// ParseEndpointFromModel parses all elements from modelEP into a Endpoint.
func ParseEndpointFromModel(modelEP *models.Endpoint) *v1.Endpoint {
	var ns, podName, containerID string
	var securityIdentity identity.NumericIdentity
	var labels []string
	if modelEP.Status != nil {
		if modelEP.Status.ExternalIdentifiers != nil {
			containerID = modelEP.Status.ExternalIdentifiers.ContainerID
			ns, podName = k8s.ParseNamespaceName(modelEP.Status.ExternalIdentifiers.PodName)
		}
		if modelEP.Status.Identity != nil {
			securityIdentity = identity.NumericIdentity(modelEP.Status.Identity.ID)
			labels = modelEP.Status.Identity.Labels
			sort.Strings(labels)
		}
	}
	ep := &v1.Endpoint{
		ID:           uint64(modelEP.ID),
		Identity:     securityIdentity,
		PodName:      podName,
		PodNamespace: ns,
		Labels:       labels,
	}

	if containerID != "" {
		ep.ContainerIDs = []string{containerID}
	}
	if modelEP.Status != nil && modelEP.Status.Networking != nil {
		// Right now we assume the endpoint will only have one IPv4 and one IPv6
		for _, ip := range modelEP.Status.Networking.Addressing {
			if ipv4 := net.ParseIP(ip.IPV4).To4(); ipv4 != nil {
				ep.IPv4 = ipv4
			}
			if ipv6 := net.ParseIP(ip.IPV6).To16(); ipv6 != nil {
				ep.IPv6 = ipv6
			}
		}
	}

	return ep
}

// ParseEndpointFromEndpointDeleteNotification returns an endpoint parsed from
// the EndpointDeleteNotification.
func ParseEndpointFromEndpointDeleteNotification(edn monitorAPI.EndpointNotification) *v1.Endpoint {
	return &v1.Endpoint{
		ID:           edn.ID,
		PodName:      edn.PodName,
		PodNamespace: edn.Namespace,
	}
}
