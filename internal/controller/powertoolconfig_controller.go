package controller

import (
	"context"
	"time"

	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	toev1alpha1 "toe/api/v1alpha1"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//+kubebuilder:rbac:groups=codriverlabs.ai.toe.run,resources=powertoolconfigs,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=codriverlabs.ai.toe.run,resources=powertoolconfigs/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=codriverlabs.ai.toe.run,resources=powertoolconfigs/finalizers,verbs=update

type PowerToolConfigReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// Reconcile is part of the main kubernetes reconciliation loop
func (r *PowerToolConfigReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the PowerToolConfig instance
	var toolConfig toev1alpha1.PowerToolConfig
	if err := r.Get(ctx, req.NamespacedName, &toolConfig); err != nil {
		logger.Error(err, "unable to fetch PowerToolConfig")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling PowerToolConfig", "name", toolConfig.Name, "tool", toolConfig.Spec.Name)

	// Update status to indicate validation
	now := metav1.Now()
	toolConfig.Status.LastValidated = &now
	toolConfig.Status.Phase = stringPtr("Ready")

	// Add condition
	condition := toev1alpha1.PowerToolConfigCondition{
		Type:               "Ready",
		Status:             "True",
		LastTransitionTime: now,
		Reason:             "ConfigurationValid",
		Message:            "PowerToolConfig is valid and ready for use",
	}

	// Update or add the condition
	toolConfig.Status.Conditions = updateCondition(toolConfig.Status.Conditions, condition)

	if err := r.Status().Update(ctx, &toolConfig); err != nil {
		logger.Error(err, "failed to update PowerToolConfig status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{RequeueAfter: 5 * time.Minute}, nil
}

// Helper function to update conditions
func updateCondition(conditions []toev1alpha1.PowerToolConfigCondition, newCondition toev1alpha1.PowerToolConfigCondition) []toev1alpha1.PowerToolConfigCondition {
	for i, condition := range conditions {
		if condition.Type == newCondition.Type {
			conditions[i] = newCondition
			return conditions
		}
	}
	return append(conditions, newCondition)
}

// Helper function to create string pointer
func stringPtr(s string) *string {
	return &s
}

// SetupWithManager sets up the controller with the Manager.
func (r *PowerToolConfigReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toev1alpha1.PowerToolConfig{}).
		Complete(r)
}
