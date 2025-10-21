package auth

import (
	"context"
	"fmt"

	authv1 "k8s.io/api/authentication/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type K8sTokenValidator struct {
	client   kubernetes.Interface
	audience string
}

func NewK8sTokenValidator(client kubernetes.Interface, audience string) *K8sTokenValidator {
	return &K8sTokenValidator{
		client:   client,
		audience: audience,
	}
}

func (v *K8sTokenValidator) ValidateToken(ctx context.Context, token string) (*authv1.UserInfo, error) {
	// Create TokenReview
	tr := &authv1.TokenReview{
		Spec: authv1.TokenReviewSpec{
			Token:     token,
			Audiences: []string{v.audience},
		},
	}

	// Submit TokenReview to API server
	result, err := v.client.AuthenticationV1().TokenReviews().Create(ctx, tr, metav1.CreateOptions{})
	if err != nil {
		return nil, fmt.Errorf("token review failed: %w", err)
	}

	if !result.Status.Authenticated {
		return nil, fmt.Errorf("token not authenticated: %v", result.Status.Error)
	}

	// Return user info for further processing
	return &result.Status.User, nil
}
