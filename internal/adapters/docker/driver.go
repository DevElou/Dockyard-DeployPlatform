package docker

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/elouan/dockyard/internal/domain"
	"github.com/elouan/dockyard/internal/ports/runtime"
)

// Driver executes deployment specs against the local Docker daemon.
// It is designed to run on the deploy-agent host where Docker is available.
type Driver struct{}

func NewDriver() *Driver {
	return &Driver{}
}

// envKeyRe matches valid POSIX environment variable names.
var envKeyRe = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]*$`)

func (d *Driver) PrepareDeployment(ctx context.Context, spec runtime.DeploymentSpec) error {
	ref := imageRef(spec.Image)
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "pull", ref)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker: pull %s: %w: %s", ref, err, stderr.String())
	}
	return nil
}

func (d *Driver) ApplyRelease(ctx context.Context, spec runtime.DeploymentSpec) (runtime.DeploymentResult, error) {
	name := containerName(spec)
	ref := imageRef(spec.Image)

	// Remove any existing container (best-effort; ignore errors)
	_ = exec.CommandContext(ctx, "docker", "rm", "-f", name).Run()

	args := []string{
		"run", "-d",
		"--name", name,
		"--restart", "unless-stopped",
		"--label", "dockyard.deployment=" + spec.DeploymentID,
		"--label", "dockyard.project=" + spec.ProjectSlug,
		"--label", "dockyard.release=" + spec.ReleaseID,
	}

	for _, env := range spec.Environment {
		if !envKeyRe.MatchString(env.Key) {
			return runtime.DeploymentResult{
				DeploymentID: spec.DeploymentID,
				Status:       domain.DeploymentStatusFailed,
				Message:      fmt.Sprintf("invalid environment variable key: %q", env.Key),
			}, fmt.Errorf("docker: invalid env key %q", env.Key)
		}
		args = append(args, "-e", env.Key+"="+env.Value)
	}

	if spec.Service.InternalPort > 0 {
		args = append(args, "-p", fmt.Sprintf("%d", spec.Service.InternalPort))
	}

	args = append(args, ref)

	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stderr = &stderr
	out, err := cmd.Output()
	if err != nil {
		return runtime.DeploymentResult{
			DeploymentID: spec.DeploymentID,
			Status:       domain.DeploymentStatusFailed,
			Message:      stderr.String(),
		}, fmt.Errorf("docker: run container %s: %w: %s", name, err, stderr.String())
	}

	containerID := strings.TrimSpace(string(out))
	return runtime.DeploymentResult{
		DeploymentID: spec.DeploymentID,
		Status:       domain.DeploymentStatusDeploying,
		ContainerID:  containerID,
		StartedAt:    time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (d *Driver) CheckHealth(ctx context.Context, deploymentID string) (runtime.DeploymentResult, error) {
	return d.checkHealthByLabel(ctx, deploymentID)
}

func (d *Driver) checkHealthByLabel(ctx context.Context, deploymentID string) (runtime.DeploymentResult, error) {
	result := runtime.DeploymentResult{DeploymentID: deploymentID}

	out, err := exec.CommandContext(ctx, "docker", "ps", "-a",
		"--filter", "label=dockyard.deployment="+deploymentID,
		"--format", "{{.ID}}:{{.Status}}",
	).Output()
	if err != nil {
		return result, fmt.Errorf("docker: list containers for deployment %s: %w", deploymentID, err)
	}

	line := strings.TrimSpace(string(out))
	if line == "" {
		result.Status = domain.DeploymentStatusFailed
		result.Message = "no container found for deployment"
		return result, nil
	}

	parts := strings.SplitN(line, ":", 2)
	result.ContainerID = parts[0]

	if len(parts) > 1 && strings.HasPrefix(strings.ToLower(parts[1]), "up") {
		result.Status = domain.DeploymentStatusHealthy
	} else {
		result.Status = domain.DeploymentStatusFailed
		result.Message = "container not running"
	}
	return result, nil
}

// GetContainerLogs returns the last `tail` log lines from the container for
// the given deployment. The container is located by the
// `dockyard.deployment=<id>` label set by ApplyRelease.
func (d *Driver) GetContainerLogs(ctx context.Context, deploymentID string, tail int) (runtime.ContainerLogs, error) {
	if tail <= 0 {
		tail = 300
	}

	idOut, err := exec.CommandContext(ctx, "docker", "ps", "-aq",
		"--filter", "label=dockyard.deployment="+deploymentID,
	).Output()
	if err != nil {
		return runtime.ContainerLogs{}, fmt.Errorf("docker: find containers for deployment %s: %w", deploymentID, err)
	}

	containerID := strings.TrimSpace(string(idOut))
	if containerID == "" {
		return runtime.ContainerLogs{}, fmt.Errorf("docker: no container for deployment %s", deploymentID)
	}
	if idx := strings.Index(containerID, "\n"); idx >= 0 {
		containerID = containerID[:idx]
	}

	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "logs",
		"--tail", fmt.Sprintf("%d", tail),
		containerID,
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return runtime.ContainerLogs{}, fmt.Errorf("docker: logs %s: %w: %s", containerID, err, stderr.String())
	}

	// docker logs writes both stdout and stderr; concatenate for the API caller.
	combined := stdout.String() + stderr.String()
	return runtime.ContainerLogs{ContainerID: containerID, Logs: combined}, nil
}

func (d *Driver) Rollback(ctx context.Context, deploymentID string, targetReleaseID string) (runtime.DeploymentResult, error) {
	// Rollback is handled by the orchestrator creating a new Deployment pointing
	// to the target release. The driver only cleans up the current deployment.
	if err := d.DeleteDeployment(ctx, deploymentID); err != nil {
		return runtime.DeploymentResult{}, fmt.Errorf("docker: rollback cleanup: %w", err)
	}
	return runtime.DeploymentResult{
		DeploymentID: deploymentID,
		Status:       domain.DeploymentStatusRolledBack,
		FinishedAt:   time.Now().UTC().Format(time.RFC3339),
	}, nil
}

func (d *Driver) DeleteDeployment(ctx context.Context, deploymentID string) error {
	out, err := exec.CommandContext(ctx, "docker", "ps", "-aq",
		"--filter", "label=dockyard.deployment="+deploymentID,
	).Output()
	if err != nil {
		return fmt.Errorf("docker: find containers for deployment %s: %w", deploymentID, err)
	}

	ids := strings.Fields(strings.TrimSpace(string(out)))
	if len(ids) == 0 {
		return nil
	}

	args := append([]string{"rm", "-f"}, ids...)
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", args...)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker: remove containers for deployment %s: %w: %s", deploymentID, err, stderr.String())
	}
	return nil
}

func imageRef(img runtime.ImageRef) string {
	if img.Digest != "" {
		return img.Repository + "@" + img.Digest
	}
	if img.Tag != "" {
		return img.Repository + ":" + img.Tag
	}
	return img.Repository
}

// slugRe matches valid Docker container name characters.
var slugRe = regexp.MustCompile(`[^a-zA-Z0-9_.-]`)

func containerName(spec runtime.DeploymentSpec) string {
	safe := slugRe.ReplaceAllString(spec.ProjectSlug, "-")
	suffix := spec.DeploymentID
	if len(suffix) > 8 {
		suffix = suffix[:8]
	}
	return "dockyard-" + safe + "-" + suffix
}
