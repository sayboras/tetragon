// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Cilium

package filters

import (
	"context"
	"fmt"
	"regexp"

	"github.com/cilium/tetragon/api/v1/tetragon"
	"github.com/cilium/tetragon/pkg/event"
)

const (
	processBinary = iota
	parentBinary
	ancestorBinary
)

func filterByBinaryRegex(binaryPatterns []string, level int) (FilterFunc, error) {
	var binaries []*regexp.Regexp
	for _, pattern := range binaryPatterns {
		query, err := regexp.Compile(pattern)
		if err != nil {
			return nil, fmt.Errorf("failed to compile regexp: %w", err)
		}
		binaries = append(binaries, query)
	}
	return func(ev *event.Event) bool {
		var processes []*tetragon.Process
		switch level {
		case processBinary:
			processes = append(processes, GetProcess(ev))
		case parentBinary:
			processes = append(processes, GetParent(ev))
		case ancestorBinary:
			processes = GetAncestors(ev)
		}
		if len(processes) == 0 || processes[0] == nil {
			return false
		}
		for _, process := range processes {
			for _, binary := range binaries {
				if binary.MatchString(process.Binary) {
					return true
				}
			}
		}
		return false
	}, nil
}

type BinaryRegexFilter struct{}

func (f *BinaryRegexFilter) OnBuildFilter(_ context.Context, ff *tetragon.Filter) ([]FilterFunc, error) {
	var fs []FilterFunc
	if ff.BinaryRegex != nil {
		filters, err := filterByBinaryRegex(ff.BinaryRegex, processBinary)
		if err != nil {
			return nil, err
		}
		fs = append(fs, filters)
	}
	return fs, nil
}

type ParentBinaryRegexFilter struct{}

func (f *ParentBinaryRegexFilter) OnBuildFilter(_ context.Context, ff *tetragon.Filter) ([]FilterFunc, error) {
	var fs []FilterFunc
	if ff.ParentBinaryRegex != nil {
		filters, err := filterByBinaryRegex(ff.ParentBinaryRegex, parentBinary)
		if err != nil {
			return nil, err
		}
		fs = append(fs, filters)
	}
	return fs, nil
}

type AncestorBinaryRegexFilter struct{}

func (f *AncestorBinaryRegexFilter) OnBuildFilter(_ context.Context, ff *tetragon.Filter) ([]FilterFunc, error) {
	var fs []FilterFunc
	if ff.AncestorBinaryRegex != nil {
		if err := CheckAncestorsEnabled(ff.EventSet); err != nil {
			return nil, err
		}

		filters, err := filterByBinaryRegex(ff.AncestorBinaryRegex, ancestorBinary)
		if err != nil {
			return nil, err
		}
		fs = append(fs, filters)
	}
	return fs, nil
}
