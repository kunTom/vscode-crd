/*
Copyright 2023.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"strconv"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/go-logr/logr"
	vscodev1 "github.com/kunTom/vscode-crd/api/v1"
	"github.com/kunTom/vscode-crd/controllers/constants"
	repo "github.com/kunTom/vscode-crd/repo"
)

// VscodeOnlineReconciler reconciles a VscodeOnline object
type VscodeOnlineReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

//+kubebuilder:rbac:groups=vscode.daocloud.io,resources=vscodeonlines,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=vscode.daocloud.io,resources=vscodeonlines/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=vscode.daocloud.io,resources=vscodeonlines/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the VscodeOnline object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.14.1/pkg/reconcile
func (r *VscodeOnlineReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := r.Log.WithValues("VscodeOnline", req.NamespacedName)
	vscode := &vscodev1.VscodeOnline{}
	err := r.Get(ctx, req.NamespacedName, vscode)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			log.Info("VscodeOnline resource not found. Ignoring since object must be deleted")
			return ctrl.Result{}, nil
		}
		// Error reading the object - requeue the request.
		log.Error(err, "Failed to get VscodeOnline")
		return ctrl.Result{}, err
	}

	vscodeDeploy, err := r.createOrGetPod(ctx, vscode, log)

	if err != nil {
		return ctrl.Result{}, err
	}

	if vscodeDeploy == nil {
		return ctrl.Result{Requeue: true}, nil
	}

	vscodeService, err := r.createOrGetService(ctx, vscode, log)

	if err != nil {
		return ctrl.Result{}, err
	}

	if vscodeService == nil {
		return ctrl.Result{Requeue: true}, nil
	}

	pvc, err := r.createOrGetPVC(ctx, vscode, vscode.Name, log, r.pvcForPlugins)
	if err != nil {
		return ctrl.Result{}, err
	}

	if pvc == nil {
		return ctrl.Result{Requeue: true}, nil
	}
	pvc, err = r.createOrGetPVC(ctx, vscode, vscode.Spec.ProjectName, log, r.pvcForCodeSpace)
	if err != nil {
		return ctrl.Result{}, err
	}

	if pvc == nil {
		return ctrl.Result{Requeue: true}, nil
	}
	pvc, err = r.createOrGetPVC(ctx, vscode, constants.VSCODE_MAVEN_SPACE_PVC, log, r.pvcForMaven)
	if err != nil {
		return ctrl.Result{}, err
	}

	if pvc == nil {
		return ctrl.Result{Requeue: true}, nil
	}

	annotations := vscodeDeploy.ObjectMeta.GetAnnotations()
	if annotations == nil {
		log.Error(err, "Failed to get git info")
		return ctrl.Result{}, err
	}

	if annotations["repo"] != vscode.Spec.Repo {
		//vscodeDeploy.Spec.Template.Spec.Containers[0].Command = r.getCommandInfo(vscode)
		vscodeDeploy.Spec.Containers[0].Command = repo.GetCommandInfo(vscode)
		err = r.Update(ctx, vscodeDeploy)
		if err != nil {
			log.Error(err, "Failed to update Pod", "Pod.Namespace", vscodeDeploy.Namespace, "Pod.Name", vscodeDeploy.Name)
			return ctrl.Result{}, err
		}
		// Spec updated - return and requeue
		return ctrl.Result{Requeue: true}, nil
	}

	svcPort := vscodeService.Spec.Ports[0].NodePort
	vscode.Status.NodePort = strconv.Itoa(int(svcPort))
	err = r.Status().Update(ctx, vscode)
	if err != nil {
		log.Error(err, "Failed to update vscode status")
		return ctrl.Result{}, err
	}
	return ctrl.Result{}, nil
}

func (r *VscodeOnlineReconciler) createOrGetPVC(ctx context.Context, vscode *vscodev1.VscodeOnline, pvcName string, log logr.Logger, creatPvc func(vscode *vscodev1.VscodeOnline) *corev1.PersistentVolumeClaim) (*corev1.PersistentVolumeClaim, error) {
	pvc := &corev1.PersistentVolumeClaim{}
	err := r.Get(ctx, types.NamespacedName{Name: pvcName, Namespace: vscode.Namespace}, pvc)
	if err != nil && errors.IsNotFound(err) {
		pvc := creatPvc(vscode)
		log.Info("Creating a new PersistentVolumeClaim", "PersistentVolumeClaim.Namespace", pvc.Namespace, "PersistentVolumeClaim.Name", pvc.Name)
		err = r.Create(ctx, pvc)
		if err != nil {
			log.Error(err, "Failed to create new PersistentVolumeClaim", "PersistentVolumeClaim.Namespace", pvc.Namespace, "PersistentVolumeClaim.Name", pvc.Name)
			return nil, err
		}
		return nil, nil
	} else if err != nil {
		log.Error(err, "Failed to get Deployment")
		return nil, err
	}
	return pvc, nil
}

func (r *VscodeOnlineReconciler) createOrGetService(ctx context.Context, vscode *vscodev1.VscodeOnline, log logr.Logger) (*corev1.Service, error) {
	vscodeService := &corev1.Service{}
	err := r.Get(ctx, types.NamespacedName{Name: vscode.Name, Namespace: vscode.Namespace}, vscodeService)
	if err != nil && errors.IsNotFound(err) {
		svc := r.serviceForVscodeOnline(vscode)
		log.Info("Creating a new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
		err = r.Create(ctx, svc)
		if err != nil {
			log.Error(err, "Failed to create new Service", "Service.Namespace", svc.Namespace, "Service.Name", svc.Name)
			return nil, err
		}
		return nil, nil
	} else if err != nil {
		log.Error(err, "Failed to get Service")
		return nil, err
	}
	return vscodeService, nil
}

func (r *VscodeOnlineReconciler) createOrGetPod(ctx context.Context, vscode *vscodev1.VscodeOnline, log logr.Logger) (*corev1.Pod, error) {
	vscodeDeploy := &corev1.Pod{}
	err := r.Get(ctx, types.NamespacedName{Name: vscode.Name, Namespace: vscode.Namespace}, vscodeDeploy)
	if err != nil && errors.IsNotFound(err) {
		pod := r.podForVscodeOnline(vscode)
		log.Info("Creating a new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
		err = r.Create(ctx, pod)
		if err != nil {
			log.Error(err, "Failed to create new Pod", "Pod.Namespace", pod.Namespace, "Pod.Name", pod.Name)
			return nil, err
		}
		return nil, nil
	} else if err != nil {
		log.Error(err, "Failed to get Pod")
		return nil, err
	}
	return vscodeDeploy, nil
}

func (r *VscodeOnlineReconciler) serviceForVscodeOnline(vscode *vscodev1.VscodeOnline) *corev1.Service {
	svc := repo.ServiceForVscodeOnline(vscode)
	ctrl.SetControllerReference(vscode, svc, r.Scheme)
	return svc
}

func (r *VscodeOnlineReconciler) pvcForPlugins(vscode *vscodev1.VscodeOnline) *corev1.PersistentVolumeClaim {
	pvc := repo.PvcForVscodeOnline(vscode.Name, constants.VSCODE_JAVA_PLUGIN_PVC, vscode.Namespace, "10Gi")
	ctrl.SetControllerReference(vscode, pvc, r.Scheme)
	return pvc
}

func (r *VscodeOnlineReconciler) pvcForCodeSpace(vscode *vscodev1.VscodeOnline) *corev1.PersistentVolumeClaim {
	pvc := repo.PvcForVscodeOnline(vscode.Name, vscode.Spec.ProjectName, vscode.Namespace, "10Gi")
	ctrl.SetControllerReference(vscode, pvc, r.Scheme)
	return pvc
}

func (r *VscodeOnlineReconciler) pvcForMaven(vscode *vscodev1.VscodeOnline) *corev1.PersistentVolumeClaim {
	pvc := repo.PvcForVscodeOnline(vscode.Name, constants.VSCODE_MAVEN_SPACE_PVC, vscode.Namespace, "10Gi")
	ctrl.SetControllerReference(vscode, pvc, r.Scheme)
	return pvc
}

func (r *VscodeOnlineReconciler) podForVscodeOnline(vscode *vscodev1.VscodeOnline) *corev1.Pod {
	pod := repo.PodForVscodeOnline(vscode)
	ctrl.SetControllerReference(vscode, pod, r.Scheme)
	return pod
}

// SetupWithManager sets up the controller with the Manager.
func (r *VscodeOnlineReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&vscodev1.VscodeOnline{}).
		Complete(r)
}
