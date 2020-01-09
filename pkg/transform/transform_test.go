package transform

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestAllOtherCRGenYaml(t *testing.T) {
	expectedSecretYaml, err := ioutil.ReadFile("testdata/expected-CR-secret.yaml")
	require.NoError(t, err)

	testCases := []struct {
		name         string
		inputCR      interface{}
		expectedYaml []byte
	}{
		{
			name: "generate yaml from secret",
			inputCR: corev1.Secret{
				TypeMeta: metav1.TypeMeta{
					APIVersion: "v1",
					Kind:       "Secret",
				},
				Data: map[string][]byte{
					"clientSecret": []byte("some-value"),
				},
				ObjectMeta: metav1.ObjectMeta{
					Name:      "literal-secret",
					Namespace: "openshift-config",
				},
				Type: "Opaque",
			},
			expectedYaml: expectedSecretYaml,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			manifest, err := GenYAML(tc.inputCR)
			require.NoError(t, err)
			assert.Equal(t, tc.expectedYaml, manifest)
		})
	}
}
