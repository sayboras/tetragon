//go:build !ignore_autogenerated

// SPDX-License-Identifier: Apache-2.0
// Copyright Authors of Tetragon

// Code generated by controller-gen. DO NOT EDIT.

package labels

import ()

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *Requirement) DeepCopyInto(out *Requirement) {
	*out = *in
	if in.strValues != nil {
		in, out := &in.strValues, &out.strValues
		*out = make([]string, len(*in))
		copy(*out, *in)
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new Requirement.
func (in *Requirement) DeepCopy() *Requirement {
	if in == nil {
		return nil
	}
	out := new(Requirement)
	in.DeepCopyInto(out)
	return out
}
