package uploadpack

import (
	"context"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command/githttp"

	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command/shared/accessverifier"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command/shared/customaction"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/command/shared/disallowedcommand"
	"gitlab.com/gitlab-org/gitlab-shell/v14/internal/config"
)

type Command struct {
	Config     *config.Config
	Args       *commandargs.Shell
	ReadWriter *readwriter.ReadWriter
}

func (c *Command) Execute(ctx context.Context) (context.Context, error) {
	args := c.Args.SshArgs
	if len(args) != 2 {
		return ctx, disallowedcommand.Error
	}

	repo := args[1]
	response, err := c.verifyAccess(ctx, repo)
	if err != nil {
		return ctx, err
	}

	logData := command.NewLogData(
		response.Gitaly.Repo.GlProjectPath,
		response.Username,
	)
	ctxWithLogData := context.WithValue(ctx, "logData", logData)

	if response.IsCustomAction() {
		if response.Payload.Data.GeoProxyFetchDirectToPrimary {
			cmd := githttp.PullCommand{
				Config:     c.Config,
				ReadWriter: c.ReadWriter,
				Response:   response,
			}

			return ctxWithLogData, cmd.Execute(ctx)
		}

		customAction := customaction.Command{
			Config:     c.Config,
			ReadWriter: c.ReadWriter,
			EOFSent:    false,
		}
		return ctxWithLogData, customAction.Execute(ctx, response)
	}

	return ctxWithLogData, c.performGitalyCall(ctx, response)
}

func (c *Command) verifyAccess(ctx context.Context, repo string) (*accessverifier.Response, error) {
	cmd := accessverifier.Command{c.Config, c.Args, c.ReadWriter}

	return cmd.Verify(ctx, c.Args.CommandType, repo)
}
