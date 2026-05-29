package domain

import (
	"errors"
	"strings"
	"time"
)

type OperationResourceType string

const (
	OperationResourceRelease    OperationResourceType = "release"
	OperationResourceDeployment OperationResourceType = "deployment"
)

type OperationLevel string

const (
	OperationLevelInfo    OperationLevel = "info"
	OperationLevelWarn    OperationLevel = "warn"
	OperationLevelError   OperationLevel = "error"
	OperationLevelSuccess OperationLevel = "success"
)

type OperationEvent struct {
	ID           string                `json:"id"`
	ResourceType OperationResourceType `json:"resourceType"`
	ResourceID   string                `json:"resourceId"`
	Phase        string                `json:"phase"`
	Level        OperationLevel        `json:"level"`
	Message      string                `json:"message"`
	Details      map[string]string     `json:"details,omitempty"`
	CreatedAt    time.Time             `json:"createdAt"`
}

func (e OperationEvent) Validate() error {
	switch e.ResourceType {
	case OperationResourceRelease, OperationResourceDeployment:
	default:
		return errors.New("operation event resource type must be release or deployment")
	}
	switch e.Level {
	case OperationLevelInfo, OperationLevelWarn, OperationLevelError, OperationLevelSuccess:
	default:
		return errors.New("operation event level must be info, warn, error, or success")
	}
	if strings.TrimSpace(e.ResourceID) == "" {
		return errors.New("operation event resource ID is required")
	}
	if strings.TrimSpace(e.Phase) == "" {
		return errors.New("operation event phase is required")
	}
	if strings.TrimSpace(e.Message) == "" {
		return errors.New("operation event message is required")
	}
	return nil
}
