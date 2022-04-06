package uploadarchive

import (
	"bytes"
	"context"
	"testing"

	"github.com/stretchr/testify/require"
	"gitlab.com/gitlab-org/labkit/correlation"

	"gitlab.com/gitlab-org/gitlab-shell/client/testserver"
	"gitlab.com/gitlab-org/gitlab-shell/internal/command/commandargs"
	"gitlab.com/gitlab-org/gitlab-shell/internal/command/readwriter"
	"gitlab.com/gitlab-org/gitlab-shell/internal/config"
	"gitlab.com/gitlab-org/gitlab-shell/internal/sshenv"
	"gitlab.com/gitlab-org/gitlab-shell/internal/testhelper/requesthandlers"
)

func TestUploadArchive(t *testing.T) {
	gitalyAddress, testServer := testserver.StartGitalyServer(t)

	requests := requesthandlers.BuildAllowedWithGitalyHandlers(t, gitalyAddress)
	url := testserver.StartHttpServer(t, requests)

	output := &bytes.Buffer{}
	input := &bytes.Buffer{}

	userId := "1"
	repo := "group/repo"

	env := sshenv.Env{
		IsSSHConnection: true,
		OriginalCommand: "git-upload-archive " + repo,
		RemoteAddr:      "127.0.0.1",
	}

	args := &commandargs.Shell{
		GitlabKeyId: userId,
		CommandType: commandargs.UploadArchive,
		SshArgs:     []string{"git-upload-archive", repo},
		Env:         env,
	}

	cmd := &Command{
		Config:     &config.Config{GitlabUrl: url},
		Args:       args,
		ReadWriter: &readwriter.ReadWriter{ErrOut: output, Out: output, In: input},
	}

	ctx := correlation.ContextWithCorrelation(context.Background(), "a-correlation-id")
	ctx = correlation.ContextWithClientName(ctx, "gitlab-shell-tests")

	err := cmd.Execute(ctx)
	require.NoError(t, err)

	require.Equal(t, "UploadArchive: "+repo, output.String())

	for k, v := range map[string]string{
		"gitaly-feature-cache_invalidator":        "true",
		"gitaly-feature-inforef_uploadpack_cache": "false",
		"x-gitlab-client-name":                    "gitlab-shell-tests-git-upload-archive",
		"key_id":                                  "123",
		"user_id":                                 "1",
		"remote_ip":                               "127.0.0.1",
		"key_type":                                "key",
	} {
		actual := testServer.ReceivedMD[k]
		require.Len(t, actual, 1)
		require.Equal(t, v, actual[0])
	}
	require.Empty(t, testServer.ReceivedMD["some-other-ff"])
	require.Equal(t, testServer.ReceivedMD["x-gitlab-correlation-id"][0], "a-correlation-id")
}
