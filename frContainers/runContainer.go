package frContainers

import (
	"context"
	"fmt"
	"github.com/docker/go-connections/nat"
	"github.com/stretchr/testify/suite"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
	"time"
)

func RunLocalstack(ctx context.Context, suite suite.Suite) (testcontainers.Container, string) {
	return RunLocalstackServices(ctx, suite, "dynamodb")
}

func RunLocalstackSqs(ctx context.Context, suite suite.Suite) (testcontainers.Container, string) {
	return RunLocalstackServices(ctx, suite, "sqs")
}

func RunLocalstackServices(ctx context.Context, suite suite.Suite, services string) (testcontainers.Container, string) {
	localstackPort := nat.Port("4566")
	req := testcontainers.ContainerRequest{
		Image:        "localstack/localstack",
		ExposedPorts: []string{"4566/tcp"},
		WaitingFor:   wait.ForListeningPort(localstackPort),
		Env: map[string]string{
			"SERVICES": services,
		},
	}
	return RunContainer(ctx, suite, req, localstackPort)
}

func RunContainer(ctx context.Context, suite suite.Suite, req testcontainers.ContainerRequest, mappedPort nat.Port) (testcontainers.Container, string) {
	retries := 3
	for i := 0; i <= retries; i++ {
		localStackContainer, err := runContainer(ctx, req)
		if localStackContainer == nil || err != nil {
			fmt.Println(fmt.Sprintf("panic Docker container start failed due to: %v", err))
			if i == retries {
				panic(err)
			}
			time.Sleep(5 * time.Second)
			continue
		}

		port, err := localStackContainer.MappedPort(ctx, mappedPort)
		suite.NoError(err)

		host, err := localStackContainer.Host(ctx)
		suite.NoError(err)

		serviceUrl := fmt.Sprintf("http://%v:%v", host, port.Port())

		return localStackContainer, serviceUrl
	}

	return nil, ""
}

func runContainer(ctx context.Context, req testcontainers.ContainerRequest) (testcontainers.Container, error) {
	defer func() {
		if panicErr := recover(); panicErr != nil {
			fmt.Println(fmt.Sprintf("panic Docker container start failed due to: %v", panicErr))
		}
	}()
	localStackContainer, err := testcontainers.GenericContainer(ctx, testcontainers.GenericContainerRequest{
		ContainerRequest: req,
		Started:          true,
	})
	return localStackContainer, err
}
