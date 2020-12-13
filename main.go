package main

import (
	"context"
	"fmt"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"net/http"
	"strconv"
)

func start(w http.ResponseWriter, req *http.Request) {

	//image, port string
	query := req.URL.Query()
	imageArr := query["image"]
	portArr := query["port"]

	if len(imageArr) == 0 && len(portArr) == 0 {
		w.WriteHeader(400)
		return
	}
	image := "cr.yandex/crpj10ofcug9f7m6kt1u/" + imageArr[0]
	port := portArr[0]
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	cli.NegotiateAPIVersion(ctx)

	hostBinding := nat.PortBinding{
		HostIP:   "",
		HostPort: port,
	}
	containerPort, err := nat.NewPort("tcp", "8090")
	if err != nil {
		panic("Unable to get the port")
	}
	portBinding := nat.PortMap{containerPort: []nat.PortBinding{hostBinding}}

	// Создаем контейнер с заданной конфигурацией
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: image,
		Cmd:   nil,
		Tty:   false,
	}, &container.HostConfig{
		PortBindings: portBinding,
	}, nil, nil, "")
	if err != nil {
		panic(err)
	}

	// Запускаем контейнер
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprint(w, "OK, started")
	}
}

func stop(w http.ResponseWriter, req *http.Request) {

	query := req.URL.Query()
	portArr := query["port"]

	if len(portArr) == 0 {
		w.WriteHeader(400)
		return
	}
	port := portArr[0]

	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		panic(err)
	}

	ctx := context.Background()
	cli.NegotiateAPIVersion(ctx)

	containers, err := cli.ContainerList(ctx, types.ContainerListOptions{All: false})
	if err != nil {
		panic(err)
	}

	result := ""
	for _, container := range containers {
		fmt.Printf("%v \n\n will stop %v %v\n\n", container, container.Ports[0].PublicPort, container.Image)
		if strconv.Itoa(int(container.Ports[0].PublicPort)) == port {
			result = container.ID[:10]
		}
	}

	if err := cli.ContainerStop(ctx, result, nil); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	} else {
		fmt.Fprint(w, "Stopped")
	}
}

func main() {
	http.HandleFunc("/deploy/start", start)
	http.HandleFunc("/deploy/stop", stop)
	http.ListenAndServe(":3000", nil)
}
