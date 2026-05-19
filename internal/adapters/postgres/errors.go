package postgres

import "errors"

var (
	ErrProjectSlugExists   = errors.New("project slug already exists")
	ErrProjectNotFound     = errors.New("project not found")

	ErrRuntimeTargetNotFound    = errors.New("runtime target not found")
	ErrRuntimeTargetSlugExists  = errors.New("runtime target slug already exists")

	ErrProjectRuntimeTargetExists = errors.New("runtime target already associated with project")

	ErrReleaseNotFound      = errors.New("release not found")
	ErrReleaseVersionExists = errors.New("release version already exists for this project")
	ErrReleaseDigestExists  = errors.New("release image digest already exists for this project")

	ErrDeploymentNotFound = errors.New("deployment not found")

	ErrDomainNotFound       = errors.New("domain not found")
	ErrDomainHostnameExists = errors.New("domain hostname already exists")
)
