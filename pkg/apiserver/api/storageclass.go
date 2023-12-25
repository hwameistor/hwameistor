package api

import v1 "k8s.io/api/storage/v1"

type StorageClass struct {
	v1.StorageClass
}

type StorageClassItemsList struct {
	StorageClasses []*StorageClass `json:"storage_classes"`
}

type StorageClassList struct {
	StorageClasses []*StorageClass `json:"storage_classes"`
	Page           *Pagination     `json:"pagination,omitempty"`
}
