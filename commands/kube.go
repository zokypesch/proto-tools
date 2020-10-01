package commands

import (
	"github.com/jinzhu/gorm"
)

// Deployment for type mapping kube
type Deployment struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   Metadata `yaml:"metadata"`
	Spec       Spec     `yaml:"spec"`
}

// Metadata for metadata kube
type Metadata struct {
	Labels    Labels `yaml:"labels"`
	Name      string `yaml:"name"`
	Namespace string `yaml:"namespace"`
}

// Labels for labeling
type Labels struct {
	Environment string `yaml:"environment"`
	Tier        string `yaml:"tier"`
	Version     string `yaml:"version"`
}

// Spec for spec
type Spec struct {
	ProgressDeadlineSeconds int `yaml:"progressDeadlineSeconds"`
	//   selector:
	//     matchLabels:
	//       app: batch
	//   strategy:
	//     rollingUpdate:
	//       maxSurge: 1
	//       maxUnavailable: 0
	//     type: RollingUpdate
	//   template:
	//     metadata:
	//       creationTimestamp: null
	//       labels:
	//         app: batch
	//         environment: production
	//         tier: backend
	//         version: v1
}

// Selector for spec
type Selector struct {
}

// KubeSvc for deploy struct
type KubeSvc struct {
	ID          int64
	Code        string
	Environment string
	Name        string
	Host        string
	Version     string
	MinScale    int
	MaxScale    int
	KubeEnv     []KubeEnv
	KubeHistory KubeHistory
}

// KubeEnv for env
type KubeEnv struct {
	ID      int64
	Code    string
	Version string
	Name    string
	Val     string
}

// KubeDigest for digest
type KubeDigest struct {
	ID      int64
	Name    string
	Digest  string
	Comment string
}

// KubeHistory for digest
type KubeHistory struct {
	ID     int64
	Code   string
	Digest string
}

// GetMeta for get meta data
func GetMeta(db *gorm.DB, svcName, env, version string) (*KubeSvc, error) {
	var svc KubeSvc

	err := db.Raw("SELECT * FROM kube_svc WHERE name = ? AND environment = ? AND version = ?", svcName, env, version).Scan(&svc).Error
	if err != nil {
		return nil, err
	}

	var kubeEnv []KubeEnv
	err = db.Raw("SELECT * FROM kube_env WHERE code = ?", svc.Code).Scan(&kubeEnv).Error
	if err != nil {
		return nil, err
	}

	var kubeHistory KubeHistory
	err = db.Raw("SELECT * FROM kube_history WHERE code = ? ORDER BY id DESC LIMIT 1", svc.Code).Scan(&kubeHistory).Error
	if err != nil && !gorm.IsRecordNotFoundError(err) {
		return nil, err
	}

	// log.Println("get latest of digest: ", kubeHistory.Digest)

	svc.KubeEnv = kubeEnv
	svc.KubeHistory = kubeHistory

	return &svc, nil
}
