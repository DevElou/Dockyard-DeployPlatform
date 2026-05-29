package domain_test

import (
	"testing"

	"github.com/elouan/dockyard/internal/domain"
)

func TestOperationEvent_Validate(t *testing.T) {
	valid := domain.OperationEvent{
		ResourceType: domain.OperationResourceRelease,
		ResourceID:   "00000000-0000-0000-0000-000000000001",
		Phase:        "building_image",
		Level:        domain.OperationLevelInfo,
		Message:      "building",
	}

	tests := []struct {
		name    string
		mutate  func(e *domain.OperationEvent)
		wantErr bool
	}{
		{name: "valid", mutate: func(e *domain.OperationEvent) {}, wantErr: false},
		{name: "bad resource type", mutate: func(e *domain.OperationEvent) { e.ResourceType = "service" }, wantErr: true},
		{name: "bad level", mutate: func(e *domain.OperationEvent) { e.Level = "trace" }, wantErr: true},
		{name: "blank resource id", mutate: func(e *domain.OperationEvent) { e.ResourceID = "  " }, wantErr: true},
		{name: "blank phase", mutate: func(e *domain.OperationEvent) { e.Phase = "" }, wantErr: true},
		{name: "blank message", mutate: func(e *domain.OperationEvent) { e.Message = "" }, wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ev := valid
			tt.mutate(&ev)
			err := ev.Validate()
			if tt.wantErr && err == nil {
				t.Fatal("expected validation error, got nil")
			}
			if !tt.wantErr && err != nil {
				t.Fatalf("unexpected validation error: %v", err)
			}
		})
	}
}
