package main

import (
	"bufio"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"runtime"
	"strings"

	"gopkg.in/urfave/cli.v1"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/swarm"
	"github.com/docker/docker/cli/command/formatter"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/sockets"
	"github.com/docker/go-connections/tlsconfig"
)

const (
	clientVersion = "1.25" // docker client version
)

func validate(host string) {
	//make sure we are running on a manager in swarm mode.
	//if not, exit with an error.
	isManager, err := amISwarmManager(host)
	if err != nil {
		fmt.Println("ERROR: ", err)
		os.Exit(1)
	}
	if !isManager {
		fmt.Println("ERROR: This script needs to run on a swarm manager.")
		os.Exit(1)
	}
}

// AmISwarmLeader determines if the current node is the swarm manager leader
func amISwarmManager(host string) (bool, error) {
	client, ctx := DockerClient(host)
	info, err := client.Info(ctx)

	if err != nil {
		return false, err
	}

	// inspect itself to see if i am the leader
	node, _, err := client.NodeInspectWithRaw(ctx, info.Swarm.NodeID)

	if err != nil {
		return false, err
	}

	if node.ManagerStatus == nil {
		return false, nil
	}

	if node.Spec.Role == "manager" {
		return true, nil
	}

	return false, errors.New("Node is NOT a manager")

}

// Ask the user to confirm.
func askToConfirm(s string) bool {
	reader := bufio.NewReader(os.Stdin)

	for {
		fmt.Printf("%s [y/N]: ", s)

		response, err := reader.ReadString('\n')
		if err != nil {
			log.Fatal(err)
		}

		response = strings.TrimSpace(response)
		if response == "y" {
			return true
		} else if response == "N" {
			return false
		}
	}
}

func verifyOK(flag bool) bool {
	if !flag {
		ok := askToConfirm("Are you sure, you want to do this?")
		if !ok {
			return false
		}
	}
	return true
}

func DockerClient(host string) (client.APIClient, context.Context) {
	// get the docker client
	tlsOptions := tlsconfig.Options{}
	ctx := context.Background()
	dockerClient, err := NewDockerClient(host, &tlsOptions)
	if err != nil {
		panic(err)
	}
	return dockerClient, ctx
}

// NewDockerClient creates a new API client.
func NewDockerClient(host string, tls *tlsconfig.Options) (client.APIClient, error) {
	tlsOptions := tls
	if tls.KeyFile == "" || tls.CAFile == "" || tls.CertFile == "" {
		// The api doesn't like it when you pass in not nil but with zero field values...
		tlsOptions = nil
	}
	customHeaders := map[string]string{
		"User-Agent": clientUserAgent(),
	}
	verStr := clientVersion
	if tmpStr := os.Getenv("DOCKER_API_VERSION"); tmpStr != "" {
		verStr = tmpStr
	}
	httpClient, err := newHTTPClient(host, tlsOptions)
	if err != nil {
		return &client.Client{}, err
	}
	return client.NewClient(host, verStr, httpClient, customHeaders)
}

func newHTTPClient(host string, tlsOptions *tlsconfig.Options) (*http.Client, error) {

	var config *tls.Config
	var err error

	if tlsOptions != nil {
		config, err = tlsconfig.Client(*tlsOptions)
		if err != nil {
			return nil, err
		}
	}
	tr := &http.Transport{
		TLSClientConfig: config,
	}
	proto, addr, _, err := client.ParseHost(host)
	if err != nil {
		return nil, err
	}
	sockets.ConfigureTransport(tr, proto, addr)
	return &http.Client{
		Transport: tr,
	}, nil
}

func clientUserAgent() string {
	// This is the client UserAgent, we use when connecting to docker.
	return fmt.Sprintf("Docker-Client/%s (%s)", clientVersion, runtime.GOOS)
}

func SwarmNodes(host string) []swarm.Node {
	cli, ctx := DockerClient(host)

	// get the list of swarm nodes
	nodes, err := cli.NodeList(ctx, types.NodeListOptions{})
	if err != nil {
		panic(err)
	}
	return nodes
}

func volumePrune(host string) (spaceReclaimed uint64, err error) {
	var output string
	cli, ctx := DockerClient(host)
	report, err := cli.VolumesPrune(ctx, types.VolumesPruneConfig{})
	if err != nil {
		return
	}

	if len(report.VolumesDeleted) > 0 {
		output += "     Deleted Volumes:\n"
		for _, id := range report.VolumesDeleted {
			output += id + "\n"
		}
		spaceReclaimed = report.SpaceReclaimed
	} else {
		output += "\n    No Volumes Deleted\n"
	}
	fmt.Println(output)
	return

}

func containerPrune(host string) (spaceReclaimed uint64, err error) {
	var output string
	cli, ctx := DockerClient(host)
	report, err := cli.ContainersPrune(ctx, types.ContainersPruneConfig{})
	if err != nil {
		return
	}
	if len(report.ContainersDeleted) > 0 {
		output += "\n    Deleted Containers:"
		for _, id := range report.ContainersDeleted {
			output += "        " + id + "\n"
		}
		spaceReclaimed = report.SpaceReclaimed
	} else {
		output += "\n    No Containers Deleted\n"
	}
	fmt.Println(output)
	return

}

func imagePrune(host string, all bool) (spaceReclaimed uint64, err error) {
	var output string
	cli, ctx := DockerClient(host)
	report, err := cli.ImagesPrune(ctx, types.ImagesPruneConfig{
		DanglingOnly: !all,
	})
	if err != nil {
		return
	}

	if len(report.ImagesDeleted) > 0 {
		output += "\n    Deleted Images:\n"
		for _, st := range report.ImagesDeleted {
			if st.Untagged != "" {
				output += fmt.Sprintln("        untagged:", st.Untagged)
			} else {
				output += fmt.Sprintln("        deleted:", st.Deleted)
			}
		}
		spaceReclaimed += report.SpaceReclaimed
	} else {
		output += "\n    No Images Deleted\n"
	}
	fmt.Println(output)
	return
}

func networkPrune(host string) (err error) {
	var output string
	cli, ctx := DockerClient(host)
	report, err := cli.NetworksPrune(ctx, types.NetworksPruneConfig{})
	if err != nil {
		return
	}

	if len(report.NetworksDeleted) > 0 {
		output += "    Deleted Networks:\n"
		for _, id := range report.NetworksDeleted {
			output += "        " + id + "\n"
		}
	} else {
		output += "\n    No Networks Deleted\n"
	}
	fmt.Println(output)
	return
}

func df(host string, verbose bool, c *cli.Context) error {
	cli, ctx := DockerClient(host)
	du, err := cli.DiskUsage(ctx)
	if err != nil {
		return err
	}

	duCtx := formatter.DiskUsageContext{
		Context: formatter.Context{
			Output: c.App.Writer,
		},
		LayersSize: du.LayersSize,
		Images:     du.Images,
		Containers: du.Containers,
		Volumes:    du.Volumes,
		Verbose:    verbose,
	}

	duCtx.Write()

	return nil
}
