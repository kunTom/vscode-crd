package repo

import (
	vscodev1 "github.com/kunTom/vscode-crd/api/v1"
	corev1 "k8s.io/api/core/v1"
)

func GetCommandInfo(vscode *vscodev1.VscodeOnline) []string {
	code_space_dir := "/home/coder/code_space/" + vscode.Name
	return []string{
		"sh",
		"-c",
		"\" mkdir -p " + code_space_dir +
			" && chmod -R 777 " + code_space_dir +
			" && git clone " + vscode.Spec.Repo + " " + code_space_dir +
			" && /usr/bin/entrypoint.sh " +
			" --bind-addr 0.0.0.0:65000 " +
			" --user-data-dir " + code_space_dir +
			" --extensions-dir /home/coder/plugin \"",
	}
}

func PodForVscodeOnline(vscode *vscodev1.VscodeOnline) *corev1.Pod {
	meta := getObjectMeta(vscode.Name, vscode.Namespace)
	annotations := map[string]string{
		"repo": vscode.Spec.Repo,
	}
	meta.SetAnnotations(annotations)
	pod := &corev1.Pod{
		ObjectMeta: meta,
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Image:   vscode.Spec.Image,
				Command: GetCommandInfo(vscode),
				Env: []corev1.EnvVar{{
					Name:  "PASSWORD",
					Value: vscode.Spec.LoginPassword,
				}},
				Ports: []corev1.ContainerPort{{
					ContainerPort: 65000,
					Name:          vscode.Name,
				}},
			}},
		},
	}
	return pod
}
