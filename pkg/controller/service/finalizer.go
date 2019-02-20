package service

import (
	"context"

	"github.com/kubesphere/porter/pkg/controller/constant"
	"github.com/kubesphere/porter/pkg/util"
	corev1 "k8s.io/api/core/v1"
)

func (r *ReconcileService) useFinalizerIfNeeded(serv *corev1.Service) (bool, error) {
	if serv.ObjectMeta.DeletionTimestamp.IsZero() {
		// The object is not being deleted, so if it does not have our finalizer,
		// then lets add the finalizer and update the object.
		if !util.ContainsString(serv.ObjectMeta.Finalizers, constant.FinalizerName) {
			serv.ObjectMeta.Finalizers = append(serv.ObjectMeta.Finalizers, constant.FinalizerName)
			if err := r.Update(context.Background(), serv); err != nil {
				return false, err
			}
			log.Info("Append Finalizer to service", "ServiceName", serv.Name, "Namespace", serv.Namespace)
		}
	} else {
		// The object is being deleted
		if util.ContainsString(serv.ObjectMeta.Finalizers, constant.FinalizerName) {
			// our finalizer is present, so lets handle our external dependency
			if err := r.deleteLB(serv); err != nil {
				// if fail to delete the external dependency here, return with error
				// so that it can be retried
				return false, err
			}

			// remove our finalizer from the list and update it.
			serv.ObjectMeta.Finalizers = util.RemoveString(serv.ObjectMeta.Finalizers, constant.FinalizerName)
			if err := r.Update(context.Background(), serv); err != nil {
				return false, err
			}
			log.Info("Remove Finalizer before service deleted", "ServiceName", serv.Name, "Namespace", serv.Namespace)
			return true, nil
		}
	}
	return false, nil
}