package arc

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/tools/clientcmd"

	imagev1 "github.com/openshift/client-go/image/clientset/versioned/typed/image/v1"
)

// OpenShiftClient communicates with the OpenShift API
type OpenShiftClient struct {
	Host   string
	images *imagev1.ImageV1Client
}

// NewOpenShiftClient returns a new, configured OpenShiftClient for the given
// kubeconfig
func NewOpenShiftClient(kubeconfigPath string) (*OpenShiftClient, error) {
	config, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, err
	}
	imageClient, err := imagev1.NewForConfig(config)
	if err != nil {
		return nil, err
	}

	return &OpenShiftClient{
		Host:   config.Host,
		images: imageClient,
	}, nil
}

// ImageMap returns a map that maps all known Image digests to their docker
// image reference
//
// Example:
// sha256:a50569906f037e14dbbe2cbf6e694730638681aca224849b686904d056233c62
// maps to
// docker.io/tnozicka/openshift-acme@sha256:a50569906f037e14dbbe2cbf6e694730638681aca224849b686904d056233c62
func (os OpenShiftClient) ImageMap() (map[string]string, error) {
	images, err := os.images.Images().List(metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	imageNames := make(map[string]string, len(images.Items))
	for i := range images.Items {
		name := images.Items[i].Name
		dockerRef := images.Items[i].DockerImageReference
		imageNames[name] = dockerRef
	}

	return imageNames, nil
}
