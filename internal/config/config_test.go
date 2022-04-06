package config

import (
	"os"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/require"

	"gitlab.com/gitlab-org/gitlab-shell/client/testserver"
	"gitlab.com/gitlab-org/gitlab-shell/internal/testhelper"
)

func TestConfigApplyGlobalState(t *testing.T) {
	t.Cleanup(testhelper.TempEnv(map[string]string{"SSL_CERT_DIR": "unmodified"}))

	config := &Config{SslCertDir: ""}
	config.ApplyGlobalState()

	require.Equal(t, "unmodified", os.Getenv("SSL_CERT_DIR"))

	config.SslCertDir = "foo"
	config.ApplyGlobalState()

	require.Equal(t, "foo", os.Getenv("SSL_CERT_DIR"))
}

func TestCustomPrometheusMetrics(t *testing.T) {
	url := testserver.StartHttpServer(t, []testserver.TestRequestHandler{})

	config := &Config{GitlabUrl: url}
	client, err := config.HttpClient()
	require.NoError(t, err)

	_, err = client.Get(url)
	require.NoError(t, err)

	ms, err := prometheus.DefaultGatherer.Gather()
	require.NoError(t, err)

	var actualNames []string
	for _, m := range ms[0:6] {
		actualNames = append(actualNames, m.GetName())
	}

	expectedMetricNames := []string{
		"gitlab_shell_http_in_flight_requests",
		"gitlab_shell_http_request_duration_seconds",
		"gitlab_shell_http_requests_total",
		"gitlab_shell_sshd_concurrent_limited_sessions_total",
		"gitlab_shell_sshd_connection_duration_seconds",
		"gitlab_shell_sshd_in_flight_connections",
	}

	require.Equal(t, expectedMetricNames, actualNames)
}
