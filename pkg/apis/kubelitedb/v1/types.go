package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// SQLiteInstanceSpec defines the desired state of SQLiteInstance
type SQLiteInstanceSpec struct {
	DbName   string `json:"dbName"`
	Storage  string `json:"storage"`
	Replicas int    `json:"replicas"`
}

// SQLiteInstanceStatus defines the observed state of SQLiteInstance
type SQLiteInstanceStatus struct {
	Phase string `json:"phase"`
}

// +kubebuilder:object:root=true

// SQLiteInstance is the Schema for the sqliteinstances API
type SQLiteInstance struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   SQLiteInstanceSpec   `json:"spec,omitempty"`
	Status SQLiteInstanceStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// SQLiteInstanceList contains a list of SQLiteInstance
type SQLiteInstanceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []SQLiteInstance `json:"items"`
}
