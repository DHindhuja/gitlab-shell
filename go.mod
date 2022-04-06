module gitlab.com/gitlab-org/gitlab-shell

go 1.16

require (
	github.com/grpc-ecosystem/go-grpc-middleware v1.3.0
	github.com/grpc-ecosystem/go-grpc-prometheus v1.2.0
	github.com/mattn/go-shellwords v1.0.11
	github.com/mikesmitty/edkey v0.0.0-20170222072505-3356ea4e686a
	github.com/otiai10/copy v1.4.2
	github.com/pires/go-proxyproto v0.6.1
	github.com/prometheus/client_golang v1.11.0
	github.com/stretchr/testify v1.7.0
	gitlab.com/gitlab-org/gitaly/v14 v14.9.0-rc5.0.20220329111719-51da8bc17059
	gitlab.com/gitlab-org/labkit v1.12.0
	golang.org/x/crypto v0.0.0-20210711020723-a769d52b0f97
	golang.org/x/sync v0.0.0-20210220032951-036812b2e83c
	google.golang.org/grpc v1.40.0
	gopkg.in/yaml.v2 v2.4.0
)

replace golang.org/x/crypto => gitlab.com/gitlab-org/golang-crypto v0.0.0-20220128174055-5be136049a80
