/*
Copyright 2023.

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

package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// VscodeOnlineSpec defines the desired state of VscodeOnline
type VscodeOnlineSpec struct {
	//code-server (vscode) image
	Image string `json:"image"`
	//user define login code-server password
	LoginPassword string `json:"loginPassword"`
	//download repository address
	Repo string `json:"repo"`
	//use Ingress or Nodeport for visitor code-server
	SvcType string `json:"svcType,omitempty"`
}

// VscodeOnlineStatus defines the observed state of VscodeOnline
type VscodeOnlineStatus struct {
	//svc nodeport
	NodePort string `json:"nodePort"`
}

//+kubebuilder:object:root=true
//+kubebuilder:subresource:status

// VscodeOnline is the Schema for the vscodeonlines API
type VscodeOnline struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   VscodeOnlineSpec   `json:"spec,omitempty"`
	Status VscodeOnlineStatus `json:"status,omitempty"`
}

//+kubebuilder:object:root=true

// VscodeOnlineList contains a list of VscodeOnline
type VscodeOnlineList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []VscodeOnline `json:"items"`
}

func init() {
	SchemeBuilder.Register(&VscodeOnline{}, &VscodeOnlineList{})
}
