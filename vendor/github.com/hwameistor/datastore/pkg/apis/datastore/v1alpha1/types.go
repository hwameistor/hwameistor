package v1alpha1

import "time"

const (
	DataSourceTypeMinIO   = "minio"
	DataSourceTypeAWSS3   = "aws-s3"
	DataSourceTypeNFS     = "nfs"
	DataSourceTypeFTP     = "ftp"
	DataSourceTypeUnknown = "unknown"
)

const (
	DataSourceConnectionCheckInterval = 1 * time.Minute
)

type MinIOSpec struct {
	Endpoint  string `json:"endpoint,omitempty"`
	Region    string `json:"region,omitempty"`
	Bucket    string `json:"bucket"`
	Prefix    string `json:"prefix,omitempty"`
	SecretKey string `json:"secretKey"`
	AccessKey string `json:"accessKey"`
}

type NFSSpec struct {
	Endpoint string `json:"endpoint,omitempty"`
	Export   string `json:"export"`
	// +kubebuilder:default:=.
	RootDir string `json:"rootdir"`
}

type FTPSpec struct {
	Endpoint string `json:"endpoint,omitempty"`
	// +kubebuilder:default:=/
	Dir           string `json:"dir"`
	LoginUser     string `json:"user"`
	LoginPassword string `json:"password"`
}

type HTTPSpec struct {
	Url string `json:"url"`
}

type SSHSpec struct {
	Node string `json:"node,omitempty"`
	// +kubebuilder:default:=22
	Port          int64  `json:"port"`
	Dir           string `json:"dir"`
	LoginUser     string `json:"user"`
	LoginPassword string `json:"password"`
}
