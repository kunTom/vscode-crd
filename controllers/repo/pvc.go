package repo

import (
	"github.com/kunTom/vscode-crd/controllers/constants"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func PvcForVscodeOnline(userName string, pvcName string, namespace string, capacity string) *corev1.PersistentVolumeClaim {
	lv := labelForVscode(userName)
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: getObjectMeta(pvcName, namespace),
		Spec: corev1.PersistentVolumeClaimSpec{
			AccessModes: []corev1.PersistentVolumeAccessMode{
				corev1.ReadWriteOnce,
			},
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceStorage: resource.MustParse(capacity),
				},
			},
			VolumeName: constants.DEFAULT_PV_NAME,
			Selector:   &metav1.LabelSelector{MatchLabels: lv},
		},
	}

	return pvc
}
