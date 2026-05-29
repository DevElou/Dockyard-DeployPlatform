package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/elouan/dockyard/internal/domain"
)

func TestResolveBuildPaths_Defaults(t *testing.T) {
	workDir := t.TempDir()
	writeFile(t, filepath.Join(workDir, "Dockerfile"))

	buildContext, dockerfilePath, err := resolveBuildPaths(workDir, domain.Project{
		DockerfilePath: "Dockerfile",
		BuildContext:   ".",
		RootDirectory:  ".",
	})
	if err != nil {
		t.Fatalf("resolveBuildPaths: %v", err)
	}
	if buildContext != workDir {
		t.Fatalf("build context: got %q, want %q", buildContext, workDir)
	}
	if dockerfilePath != filepath.Join(workDir, "Dockerfile") {
		t.Fatalf("dockerfile path: got %q", dockerfilePath)
	}
}

func TestResolveBuildPaths_RootDirectory(t *testing.T) {
	workDir := t.TempDir()
	appDir := filepath.Join(workDir, "apps", "api")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(appDir, "Dockerfile"))

	buildContext, dockerfilePath, err := resolveBuildPaths(workDir, domain.Project{
		DockerfilePath: "Dockerfile",
		BuildContext:   ".",
		RootDirectory:  "apps/api",
	})
	if err != nil {
		t.Fatalf("resolveBuildPaths: %v", err)
	}
	if buildContext != appDir {
		t.Fatalf("build context: got %q, want %q", buildContext, appDir)
	}
	if dockerfilePath != filepath.Join(appDir, "Dockerfile") {
		t.Fatalf("dockerfile path: got %q", dockerfilePath)
	}
}

func TestResolveBuildPaths_DockerfileRelativeToRootDirectory(t *testing.T) {
	workDir := t.TempDir()
	appDir := filepath.Join(workDir, "app")
	contextDir := filepath.Join(appDir, "src")
	if err := os.MkdirAll(contextDir, 0o755); err != nil {
		t.Fatal(err)
	}
	writeFile(t, filepath.Join(appDir, "build", "Dockerfile"))

	buildContext, dockerfilePath, err := resolveBuildPaths(workDir, domain.Project{
		DockerfilePath: "build/Dockerfile",
		BuildContext:   "src",
		RootDirectory:  "app",
	})
	if err != nil {
		t.Fatalf("resolveBuildPaths: %v", err)
	}
	if buildContext != contextDir {
		t.Fatalf("build context: got %q, want %q", buildContext, contextDir)
	}
	if dockerfilePath != filepath.Join(appDir, "build", "Dockerfile") {
		t.Fatalf("dockerfile path: got %q", dockerfilePath)
	}
}

func TestResolveBuildPaths_MissingDockerfile(t *testing.T) {
	workDir := t.TempDir()

	_, _, err := resolveBuildPaths(workDir, domain.Project{
		DockerfilePath: "Dockerfile",
		BuildContext:   ".",
		RootDirectory:  ".",
	})
	if err == nil {
		t.Fatal("expected missing dockerfile error")
	}
}

func TestSafeJoinRejectsEscapes(t *testing.T) {
	workDir := t.TempDir()
	for _, rel := range []string{"../Dockerfile", "/Dockerfile"} {
		if _, err := safeJoin(workDir, rel); err == nil {
			t.Fatalf("expected %q to be rejected", rel)
		}
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(path, []byte("FROM scratch\n"), 0o644); err != nil {
		t.Fatal(err)
	}
}
