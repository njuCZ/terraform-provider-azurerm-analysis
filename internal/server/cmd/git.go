package cmd

import (
	"context"
	"os/exec"
)

type GitRepo struct {
	GitCmdPath string
	Dir        string
}

func (repo *GitRepo) Pull(ctx context.Context) error {
	cmd := exec.CommandContext(ctx, repo.GitCmdPath, "pull")
	cmd.Dir = repo.Dir

	return cmd.Run()
}
