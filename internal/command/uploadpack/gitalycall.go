package uploadpack

import (
	"context"

	"google.golang.org/grpc"

	"gitlab.com/gitlab-org/gitaly/v14/client"
	pb "gitlab.com/gitlab-org/gitaly/v14/proto/go/gitalypb"
	"gitlab.com/gitlab-org/gitlab-shell/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/internal/gitlabnet/accessverifier"
	"gitlab.com/gitlab-org/gitlab-shell/internal/handler"
)

func (c *Command) performGitalyCall(ctx context.Context, response *accessverifier.Response) error {
	gc := handler.NewGitalyCommand(c.Config, string(commandargs.UploadPack), response)

	if response.Gitaly.UseSidechannel {
		request := &pb.SSHUploadPackWithSidechannelRequest{
			Repository:       &response.Gitaly.Repo,
			GitProtocol:      c.Args.Env.GitProtocolVersion,
			GitConfigOptions: response.GitConfigOptions,
		}

		return gc.RunGitalyCommand(ctx, func(ctx context.Context, conn *grpc.ClientConn) (int32, error) {
			ctx, cancel := gc.PrepareContext(ctx, request.Repository, c.Args.Env)
			defer cancel()

			registry := c.Config.GitalyClient.SidechannelRegistry
			rw := c.ReadWriter
			return client.UploadPackWithSidechannel(ctx, conn, registry, rw.In, rw.Out, rw.ErrOut, request)
		})
	}

	request := &pb.SSHUploadPackRequest{
		Repository:       &response.Gitaly.Repo,
		GitProtocol:      c.Args.Env.GitProtocolVersion,
		GitConfigOptions: response.GitConfigOptions,
	}

	return gc.RunGitalyCommand(ctx, func(ctx context.Context, conn *grpc.ClientConn) (int32, error) {
		ctx, cancel := gc.PrepareContext(ctx, request.Repository, c.Args.Env)
		defer cancel()

		rw := c.ReadWriter
		return client.UploadPack(ctx, conn, rw.In, rw.Out, rw.ErrOut, request)
	})
}
