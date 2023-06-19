package repo

import metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

func getObjectMeta(name string, namespace string) metav1.ObjectMeta {
	return metav1.ObjectMeta{
		Name:      name,
		Namespace: namespace,
	}
}

func labelForVscode(name string) map[string]string {
	return map[string]string{"app": "VscodeOnline", "VscodeOnline": name}
}
