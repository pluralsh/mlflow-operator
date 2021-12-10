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

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// TrackingServerSpec defines the desired state of TrackingServer
type TrackingServerSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Image string `json:"image"`

	//+kubebuilder:validation:Optional
	ImagePullSecret string `json:"imagePullSecret"`

	Replicas int32 `json:"replicas"`

	Size string `json:"size"`

	//+kubebuilder:validation:Optional
	StorageClass string `json:"storageClass"`

	//+kubebuilder:validation:Optional
	//+kubebuilder:default:="file:///mnt/mlruns/artifacts"
	DefaultArtifactRoot string `json:"defaultArtifactRoot"`

	//+kubebuilder:validation:Optional
	ServiceAccountName string `json:"serviceAccountName"`

	//+kubebuilder:validation:Optional
	ExtraPodLabels map[string]string `json:"extraPodLabels"`

	//+kubebuilder:validation:Optional
	S3endpointURL string `json:"s3endpointURL,omitempty"`

	//+kubebuilder:validation:Optional
	S3secretName string `json:"s3secretName,omitempty"`

	Network TrackingServerSpecNetworkConfig `json:"network,omitempty"`
}

type TrackingServerSpecNetworkConfig struct {
	// Name of the Istio Gateway to use for the VirtualService
	IstioGatewayName string `json:"istioGatewayName,omitempty"`

	// Name of the Istio Gateway to use for the VirtualService
	IstioGatewayNamespace string `json:"istioGatewayNamespace,omitempty"`
}

// TrackingServerStatus defines the observed state of TrackingServer
type TrackingServerStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// TrackingServer is the Schema for the trackingservers API
type TrackingServer struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   TrackingServerSpec   `json:"spec,omitempty"`
	Status TrackingServerStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// TrackingServerList contains a list of TrackingServer
type TrackingServerList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []TrackingServer `json:"items"`
}

func init() {
	SchemeBuilder.Register(&TrackingServer{}, &TrackingServerList{})
}
