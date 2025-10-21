package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	toev1alpha1 "toe/api/v1alpha1"
	"toe/pkg/collector/auth"
)

// Reconciliation timing constants
const (
	ActiveRunningInterval   = 5 * time.Second
	SetupTeardownInterval   = 15 * time.Second
	CompletedJobInterval    = 5 * time.Minute
	EphemeralStatusInterval = 3 * time.Second
)

// Output mode constants
const (
	OutputModePVC = "pvc"
)

// Phase constants
const (
	PhaseCompleted = "Completed"
)

//+kubebuilder:rbac:groups=codriverlabs.ai.toe.run,resources=powertools,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=codriverlabs.ai.toe.run,resources=powertools/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=codriverlabs.ai.toe.run,resources=powertools/finalizers,verbs=update
//+kubebuilder:rbac:groups=codriverlabs.ai.toe.run,resources=powertoolconfigs,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=pods/ephemeralcontainers,verbs=get;list;watch;update;patch
//+kubebuilder:rbac:groups="",resources=configmaps,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts,verbs=get;list;watch
//+kubebuilder:rbac:groups="",resources=serviceaccounts/token,verbs=create

type PowerToolReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	K8sClient kubernetes.Interface
}

func NewPowerToolReconciler(c client.Client, scheme *runtime.Scheme, k8sClient kubernetes.Interface) *PowerToolReconciler {
	return &PowerToolReconciler{
		Client:    c,
		Scheme:    scheme,
		K8sClient: k8sClient,
	}
}

func (r *PowerToolReconciler) getToolConfig(ctx context.Context, toolName string) (*toev1alpha1.PowerToolConfig, error) {
	// Look for PowerToolConfig in the same namespace first, then toe-system
	namespaces := []string{"toe-system", "default"}

	for _, namespace := range namespaces {
		var toolConfig toev1alpha1.PowerToolConfig
		configKey := client.ObjectKey{
			Name:      toolName + "-config",
			Namespace: namespace,
		}

		if err := r.Get(ctx, configKey, &toolConfig); err == nil {
			return &toolConfig, nil
		}
	}

	return nil, fmt.Errorf("PowerToolConfig not found for tool: %s", toolName)
}

func (r *PowerToolReconciler) getTokenDuration(ctx context.Context, collectionDuration time.Duration) time.Duration {
	logger := log.FromContext(ctx)

	// Simple calculation: collection duration + 60 seconds buffer for overhead
	buffer := 60 * time.Second
	tokenDuration := collectionDuration + buffer

	// Kubernetes minimum requirement: 10 minutes (600 seconds)
	minDuration := 10 * time.Minute
	if tokenDuration < minDuration {
		logger.Info("Token duration below minimum, using 10 minutes",
			"calculated", tokenDuration,
			"minimum", minDuration,
			"collectionDuration", collectionDuration)
		tokenDuration = minDuration
	}

	logger.Info("Token duration calculated",
		"collectionDuration", collectionDuration,
		"buffer", buffer,
		"finalTokenDuration", tokenDuration)

	return tokenDuration
}

// buildPowerToolEnvVars builds environment variables from PowerTool spec
func (r *PowerToolReconciler) buildPowerToolEnvVars(job *toev1alpha1.PowerTool, targetPod corev1.Pod) []corev1.EnvVar {
	envVars := []corev1.EnvVar{
		{Name: "PROFILER_TOOL", Value: job.Spec.Tool.Name},
		{Name: "PROFILER_DURATION", Value: job.Spec.Tool.Duration},
		{Name: "TARGET_POD_NAME", Value: targetPod.Name},
		{Name: "TARGET_NAMESPACE", Value: targetPod.Namespace},
		{Name: "OUTPUT_MODE", Value: job.Spec.Output.Mode},
	}

	// Add tool-specific arguments as environment variables
	if job.Spec.Tool.Args != nil && job.Spec.Tool.Args.Raw != nil {
		var args map[string]interface{}
		if err := json.Unmarshal(job.Spec.Tool.Args.Raw, &args); err == nil {
			for key, value := range args {
				envVars = append(envVars, corev1.EnvVar{
					Name:  fmt.Sprintf("TOOL_ARG_%s", key),
					Value: fmt.Sprintf("%v", value),
				})
			}
		}
	}

	// Add PVC path if specified
	if job.Spec.Output.Mode == "pvc" && job.Spec.Output.PVC != nil && job.Spec.Output.PVC.Path != nil {
		envVars = append(envVars, corev1.EnvVar{
			Name:  "PVC_PATH",
			Value: *job.Spec.Output.PVC.Path,
		})
	}

	return envVars
}

// findPVCVolumeName finds the volume name for a given PVC claim name in the pod
func (r *PowerToolReconciler) findPVCVolumeName(pod corev1.Pod, claimName string) string {
	for _, volume := range pod.Spec.Volumes {
		if volume.PersistentVolumeClaim != nil && volume.PersistentVolumeClaim.ClaimName == claimName {
			return volume.Name
		}
	}
	// Return a default name if not found
	return "profiling-storage"
}

// buildSecurityContext converts SecuritySpec to SecurityContext
func (r *PowerToolReconciler) buildSecurityContext(securitySpec toev1alpha1.SecuritySpec) *corev1.SecurityContext {
	securityContext := &corev1.SecurityContext{}

	if securitySpec.AllowPrivileged != nil {
		securityContext.Privileged = securitySpec.AllowPrivileged
	}

	if securitySpec.Capabilities != nil {
		capabilities := &corev1.Capabilities{}

		if securitySpec.Capabilities.Add != nil {
			for _, cap := range securitySpec.Capabilities.Add {
				capabilities.Add = append(capabilities.Add, corev1.Capability(cap))
			}
		}

		if securitySpec.Capabilities.Drop != nil {
			for _, cap := range securitySpec.Capabilities.Drop {
				capabilities.Drop = append(capabilities.Drop, corev1.Capability(cap))
			}
		}

		securityContext.Capabilities = capabilities
	}

	return securityContext
}

func (r *PowerToolReconciler) validateNamespaceAccess(job *toev1alpha1.PowerTool, toolConfig *toev1alpha1.PowerToolConfig) error {
	// If no namespace restrictions, allow all
	if len(toolConfig.Spec.AllowedNamespaces) == 0 {
		return nil
	}

	// Check if PowerTool namespace is in allowed list
	for _, allowedNS := range toolConfig.Spec.AllowedNamespaces {
		if job.Namespace == allowedNS {
			return nil
		}
	}

	return fmt.Errorf("PowerTool namespace '%s' is not allowed for tool '%s'. Allowed namespaces: %v",
		job.Namespace, toolConfig.Spec.Name, toolConfig.Spec.AllowedNamespaces)
}

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PowerToolReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	// Fetch the PowerTool instance
	var powerTool toev1alpha1.PowerTool
	if err := r.Get(ctx, req.NamespacedName, &powerTool); err != nil {
		logger.Error(err, "unable to fetch PowerTool")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	logger.Info("Reconciling PowerTool", "name", powerTool.Name, "namespace", powerTool.Namespace)

	// Handle deletion
	if powerTool.DeletionTimestamp != nil {
		return r.handleDeletion(ctx, &powerTool)
	}

	// Initialize status if needed
	if powerTool.Status.Phase == nil {
		phase := "Pending"
		powerTool.Status.Phase = &phase
		now := metav1.Now()
		powerTool.Status.StartedAt = &now
		r.setCondition(&powerTool, toev1alpha1.PowerToolConditionReady, "False", toev1alpha1.ReasonTargetsSelected, "Initializing PowerTool")
		if err := r.Status().Update(ctx, &powerTool); err != nil {
			logger.Error(err, "unable to update PowerTool status")
			return ctrl.Result{}, err
		}
	}

	// Get tool configuration
	toolConfig, err := r.getToolConfig(ctx, powerTool.Spec.Tool.Name)
	if err != nil {
		logger.Error(err, "failed to get tool configuration")
		r.setCondition(&powerTool, toev1alpha1.PowerToolConditionFailed, "True", toev1alpha1.ReasonFailed, fmt.Sprintf("Tool configuration error: %v", err))
		if updateErr := r.Status().Update(ctx, &powerTool); updateErr != nil {
			logger.Error(updateErr, "failed to update PowerTool status")
		}
		return ctrl.Result{}, err
	}

	// Validate namespace access
	if err := r.validateNamespaceAccess(&powerTool, toolConfig); err != nil {
		logger.Error(err, "namespace access denied")
		r.setCondition(&powerTool, toev1alpha1.PowerToolConditionFailed, "True", toev1alpha1.ReasonFailed, fmt.Sprintf("Namespace access denied: %v", err))
		if updateErr := r.Status().Update(ctx, &powerTool); updateErr != nil {
			logger.Error(updateErr, "failed to update PowerTool status")
		}
		return ctrl.Result{}, err
	}

	// Get target pods
	var podList corev1.PodList
	selector, err := metav1.LabelSelectorAsSelector(powerTool.Spec.Targets.LabelSelector)
	if err != nil {
		logger.Error(err, "unable to convert label selector")
		r.setCondition(&powerTool, toev1alpha1.PowerToolConditionFailed, "True", toev1alpha1.ReasonFailed, fmt.Sprintf("Invalid label selector: %v", err))
		if updateErr := r.Status().Update(ctx, &powerTool); updateErr != nil {
			logger.Error(updateErr, "failed to update PowerTool status")
		}
		return ctrl.Result{}, err
	}

	if err := r.List(ctx, &podList, &client.ListOptions{
		Namespace:     powerTool.Namespace,
		LabelSelector: selector,
	}); err != nil {
		logger.Error(err, "unable to list target pods")
		return ctrl.Result{}, err
	}

	selectedPods := int32(len(podList.Items))
	powerTool.Status.SelectedPods = &selectedPods

	// Check for conflicts with other active PowerTools
	if conflict, conflictMsg := r.checkForConflicts(ctx, &powerTool, podList.Items); conflict {
		r.setCondition(&powerTool, toev1alpha1.PowerToolConditionConflicted, "True", toev1alpha1.ReasonConflictDetected, conflictMsg)
		phase := "Conflicted"
		powerTool.Status.Phase = &phase
		if err := r.Status().Update(ctx, &powerTool); err != nil {
			logger.Error(err, "unable to update PowerTool status")
			return ctrl.Result{}, err
		}
		return ctrl.Result{RequeueAfter: time.Minute}, nil
	}

	// Initialize ActivePods map if needed
	if powerTool.Status.ActivePods == nil {
		powerTool.Status.ActivePods = make(map[string]string)
	}

	// Process pods for profiling
	for _, pod := range podList.Items {
		containerName := fmt.Sprintf("powertool-%s-%s", powerTool.Name, string(powerTool.UID)[:8])

		// Check if we already have a container for this pod
		if existingContainer, exists := powerTool.Status.ActivePods[pod.Name]; exists {
			if r.isContainerRunning(pod, existingContainer) {
				continue // Still running
			} else {
				// Container finished, clean up
				delete(powerTool.Status.ActivePods, pod.Name)
			}
		}

		// Check if container already exists in pod spec
		containerExists := false
		for _, ec := range pod.Spec.EphemeralContainers {
			if ec.Name == containerName {
				containerExists = true
				powerTool.Status.ActivePods[pod.Name] = containerName
				break
			}
		}

		if containerExists {
			continue
		}

		// Create new ephemeral container
		if err := r.createEphemeralContainerForPod(ctx, &powerTool, toolConfig, pod, containerName); err != nil {
			logger.Error(err, "failed to create ephemeral container", "pod", pod.Name)
			continue
		}

		powerTool.Status.ActivePods[pod.Name] = containerName
	}

	// Update status based on active containers
	completedPods := selectedPods - int32(len(powerTool.Status.ActivePods))
	powerTool.Status.CompletedPods = &completedPods

	if len(powerTool.Status.ActivePods) > 0 {
		phase := "Running"
		powerTool.Status.Phase = &phase
		r.setCondition(&powerTool, toev1alpha1.PowerToolConditionRunning, "True", toev1alpha1.ReasonRunning, fmt.Sprintf("Running on %d pods", len(powerTool.Status.ActivePods)))
	} else if selectedPods > 0 {
		phase := PhaseCompleted
		powerTool.Status.Phase = &phase
		now := metav1.Now()
		powerTool.Status.FinishedAt = &now
		r.setCondition(&powerTool, toev1alpha1.PowerToolConditionCompleted, "True", toev1alpha1.ReasonCompleted, "All containers completed")
	}

	if err := r.Status().Update(ctx, &powerTool); err != nil {
		logger.Error(err, "unable to update PowerTool status")
		return ctrl.Result{}, err
	}

	// Determine requeue interval
	interval := r.getRequeueInterval(&powerTool)
	return ctrl.Result{RequeueAfter: interval}, nil
}

func (r *PowerToolReconciler) getRequeueInterval(job *toev1alpha1.PowerTool) time.Duration {
	if job.Status.Phase == nil {
		return SetupTeardownInterval
	}

	switch *job.Status.Phase {
	case "Running":
		return ActiveRunningInterval
	case "Completed", "Failed":
		return CompletedJobInterval
	default:
		return SetupTeardownInterval
	}
}

// handleDeletion handles PowerTool deletion with proper cleanup
func (r *PowerToolReconciler) handleDeletion(ctx context.Context, powerTool *toev1alpha1.PowerTool) (ctrl.Result, error) {
	logger := log.FromContext(ctx)
	logger.Info("Handling PowerTool deletion", "name", powerTool.Name)

	// Note: Ephemeral containers cannot be removed from pods once created
	// They will be cleaned up when the pod is deleted
	// We just need to ensure proper status reporting

	return ctrl.Result{}, nil
}

// setCondition sets or updates a condition in the PowerTool status
func (r *PowerToolReconciler) setCondition(powerTool *toev1alpha1.PowerTool, conditionType, status, reason, message string) {
	now := metav1.Now()

	// Find existing condition
	for i, condition := range powerTool.Status.Conditions {
		if condition.Type == conditionType {
			if condition.Status != status {
				powerTool.Status.Conditions[i].Status = status
				powerTool.Status.Conditions[i].LastTransitionTime = now
			}
			powerTool.Status.Conditions[i].Reason = reason
			powerTool.Status.Conditions[i].Message = message
			return
		}
	}

	// Add new condition
	powerTool.Status.Conditions = append(powerTool.Status.Conditions, toev1alpha1.PowerToolCondition{
		Type:               conditionType,
		Status:             status,
		LastTransitionTime: now,
		Reason:             reason,
		Message:            message,
	})
}

// checkForConflicts checks if there are conflicting PowerTools targeting the same pods
func (r *PowerToolReconciler) checkForConflicts(ctx context.Context, currentTool *toev1alpha1.PowerTool, targetPods []corev1.Pod) (bool, string) {
	var allPowerTools toev1alpha1.PowerToolList
	if err := r.List(ctx, &allPowerTools); err != nil {
		return false, ""
	}

	for _, tool := range allPowerTools.Items {
		// Skip self and completed tools
		if tool.Name == currentTool.Name || tool.Namespace != currentTool.Namespace {
			continue
		}
		if tool.Status.Phase != nil && (*tool.Status.Phase == "Completed" || *tool.Status.Phase == "Failed") {
			continue
		}

		// Check if this tool has active pods that overlap with our targets
		if tool.Status.ActivePods != nil {
			for _, targetPod := range targetPods {
				if _, exists := tool.Status.ActivePods[targetPod.Name]; exists {
					return true, fmt.Sprintf("Pod %s is already being profiled by PowerTool %s", targetPod.Name, tool.Name)
				}
			}
		}
	}

	return false, ""
}

// isContainerRunning checks if the specified ephemeral container is still running
func (r *PowerToolReconciler) isContainerRunning(pod corev1.Pod, containerName string) bool {
	// Check if container exists in ephemeral containers
	for _, ec := range pod.Spec.EphemeralContainers {
		if ec.Name == containerName {
			// Container exists, check its status
			for _, status := range pod.Status.EphemeralContainerStatuses {
				if status.Name == containerName {
					return status.State.Running != nil
				}
			}
			// Container exists but no status yet, assume running
			return true
		}
	}
	return false
}

// createEphemeralContainerForPod creates an ephemeral container for a specific pod
func (r *PowerToolReconciler) createEphemeralContainerForPod(ctx context.Context, powerTool *toev1alpha1.PowerTool, toolConfig *toev1alpha1.PowerToolConfig, pod corev1.Pod, containerName string) error {
	logger := log.FromContext(ctx)

	// Build environment variables
	envVars := r.buildPowerToolEnvVars(powerTool, pod)

	// Add collector configuration if specified
	if powerTool.Spec.Output.Collector != nil {
		collectionDuration, err := time.ParseDuration(powerTool.Spec.Tool.Duration)
		if err != nil {
			return fmt.Errorf("invalid duration: %w", err)
		}

		tokenDuration := r.getTokenDuration(ctx, collectionDuration)

		// Create a token manager for the collector
		collectorTokenManager := auth.NewK8sTokenManager(r.K8sClient, "toe-system", "toe-sdk-collector")
		token, err := collectorTokenManager.GenerateToken(ctx, powerTool.Name, tokenDuration)
		if err != nil {
			return fmt.Errorf("failed to generate collection token: %w", err)
		}

		envVars = append(envVars,
			corev1.EnvVar{Name: "COLLECTOR_ENDPOINT", Value: powerTool.Spec.Output.Collector.Endpoint},
			corev1.EnvVar{Name: "COLLECTOR_TOKEN", Value: token},
			corev1.EnvVar{Name: "POWERTOOL_JOB_ID", Value: powerTool.Name},
		)
	}

	// Create ephemeral container
	ec := &corev1.EphemeralContainer{
		EphemeralContainerCommon: corev1.EphemeralContainerCommon{
			Name:            containerName,
			Image:           toolConfig.Spec.Image,
			ImagePullPolicy: corev1.PullAlways,
			Env:             envVars,
			SecurityContext: r.buildSecurityContext(toolConfig.Spec.SecurityContext),
		},
	}

	// Add PVC volume mount if specified
	if powerTool.Spec.Output.Mode == OutputModePVC && powerTool.Spec.Output.PVC != nil {
		ec.VolumeMounts = []corev1.VolumeMount{
			{
				Name:      r.findPVCVolumeName(pod, powerTool.Spec.Output.PVC.ClaimName),
				MountPath: "/mnt/profiling-storage",
			},
		}
	}

	// Update pod with ephemeral container
	podCopy := pod.DeepCopy()
	podCopy.Spec.EphemeralContainers = append(podCopy.Spec.EphemeralContainers, *ec)
	if err := r.SubResource("ephemeralcontainers").Update(ctx, podCopy); err != nil {
		return fmt.Errorf("failed to add ephemeral container to pod %s: %w", pod.Name, err)
	}

	logger.Info("Successfully added ephemeral container",
		"pod", pod.Name,
		"container", containerName,
		"image", toolConfig.Spec.Image)

	return nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *PowerToolReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&toev1alpha1.PowerTool{}).
		Complete(r)
}
