package dockerregistry

import (
	"bytes"
	"context"
	"fmt"
	"os/exec"
	"strings"

	"github.com/elouan/dockyard/internal/ports/registry"
)

type Builder struct {
	registryURL string
}

func NewBuilder(registryURL string) *Builder {
	return &Builder{registryURL: registryURL}
}

func (b *Builder) BuildAndPush(ctx context.Context, req registry.BuildRequest) (registry.BuildResult, error) {
	imageTag := fmt.Sprintf("%s/%s:%s", b.registryURL, req.ProjectID, req.ReleaseVersion)

	if err := b.build(ctx, imageTag, req.DockerfilePath, req.BuildContext); err != nil {
		return registry.BuildResult{}, err
	}

	if err := b.push(ctx, imageTag); err != nil {
		return registry.BuildResult{}, err
	}

	digest, err := b.inspect(ctx, imageTag)
	if err != nil {
		return registry.BuildResult{}, err
	}

	return registry.BuildResult{
		ImageRepository: fmt.Sprintf("%s/%s", b.registryURL, req.ProjectID),
		ImageTag:        req.ReleaseVersion,
		ImageDigest:     digest,
	}, nil
}

func (b *Builder) build(ctx context.Context, tag, dockerfilePath, buildContext string) error {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "build",
		"--pull",
		"-t", tag,
		"-f", dockerfilePath,
		buildContext,
	)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker build: %w: %s", err, stderr.String())
	}
	return nil
}

func (b *Builder) push(ctx context.Context, tag string) error {
	var stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "push", tag)
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker push: %w: %s", err, stderr.String())
	}
	return nil
}

func (b *Builder) inspect(ctx context.Context, tag string) (string, error) {
	var stdout, stderr bytes.Buffer
	cmd := exec.CommandContext(ctx, "docker", "inspect",
		"--format", "{{index .RepoDigests 0}}",
		tag,
	)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		return "", fmt.Errorf("docker inspect: %w: %s", err, stderr.String())
	}

	// RepoDigest format: "registry/image@sha256:abc123..."
	raw := strings.TrimSpace(stdout.String())
	if idx := strings.Index(raw, "@"); idx != -1 {
		return raw[idx+1:], nil
	}
	return raw, nil
}
