package repo

import (
	vscodev1 "github.com/kunTom/vscode-crd/api/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
)

func ServiceForVscodeOnline(vscode *vscodev1.VscodeOnline) *corev1.Service {
	lv := labelForVscode(vscode.Name)
	svc := &corev1.Service{
		ObjectMeta: getObjectMeta(vscode.Name, vscode.Namespace),
		Spec: corev1.ServiceSpec{
			Selector: lv,
			Ports: []corev1.ServicePort{{
				Name:       vscode.Name,
				Port:       int32(65000),
				TargetPort: intstr.FromInt(65000),
				Protocol:   "tcp",
			}},
			Type: corev1.ServiceTypeNodePort,
		},
	}

	return svc
}
