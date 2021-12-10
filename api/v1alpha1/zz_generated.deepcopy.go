//go:build !ignore_autogenerated
// +build !ignore_autogenerated

/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

// Code generated by controller-gen. DO NOT EDIT.

package v1alpha1

import (
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrackingServer) DeepCopyInto(out *TrackingServer) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ObjectMeta.DeepCopyInto(&out.ObjectMeta)
	in.Spec.DeepCopyInto(&out.Spec)
	out.Status = in.Status
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrackingServer.
func (in *TrackingServer) DeepCopy() *TrackingServer {
	if in == nil {
		return nil
	}
	out := new(TrackingServer)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrackingServer) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrackingServerList) DeepCopyInto(out *TrackingServerList) {
	*out = *in
	out.TypeMeta = in.TypeMeta
	in.ListMeta.DeepCopyInto(&out.ListMeta)
	if in.Items != nil {
		in, out := &in.Items, &out.Items
		*out = make([]TrackingServer, len(*in))
		for i := range *in {
			(*in)[i].DeepCopyInto(&(*out)[i])
		}
	}
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrackingServerList.
func (in *TrackingServerList) DeepCopy() *TrackingServerList {
	if in == nil {
		return nil
	}
	out := new(TrackingServerList)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyObject is an autogenerated deepcopy function, copying the receiver, creating a new runtime.Object.
func (in *TrackingServerList) DeepCopyObject() runtime.Object {
	if c := in.DeepCopy(); c != nil {
		return c
	}
	return nil
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrackingServerSpec) DeepCopyInto(out *TrackingServerSpec) {
	*out = *in
	if in.ExtraPodLabels != nil {
		in, out := &in.ExtraPodLabels, &out.ExtraPodLabels
		*out = make(map[string]string, len(*in))
		for key, val := range *in {
			(*out)[key] = val
		}
	}
	out.Network = in.Network
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrackingServerSpec.
func (in *TrackingServerSpec) DeepCopy() *TrackingServerSpec {
	if in == nil {
		return nil
	}
	out := new(TrackingServerSpec)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrackingServerSpecNetworkConfig) DeepCopyInto(out *TrackingServerSpecNetworkConfig) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrackingServerSpecNetworkConfig.
func (in *TrackingServerSpecNetworkConfig) DeepCopy() *TrackingServerSpecNetworkConfig {
	if in == nil {
		return nil
	}
	out := new(TrackingServerSpecNetworkConfig)
	in.DeepCopyInto(out)
	return out
}

// DeepCopyInto is an autogenerated deepcopy function, copying the receiver, writing into out. in must be non-nil.
func (in *TrackingServerStatus) DeepCopyInto(out *TrackingServerStatus) {
	*out = *in
}

// DeepCopy is an autogenerated deepcopy function, copying the receiver, creating a new TrackingServerStatus.
func (in *TrackingServerStatus) DeepCopy() *TrackingServerStatus {
	if in == nil {
		return nil
	}
	out := new(TrackingServerStatus)
	in.DeepCopyInto(out)
	return out
}
