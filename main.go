package main

import (
	"fmt"
	"os"
	"time"

	"github.com/docker/go-connections/tlsconfig"
	"github.com/docker/go-units"
	"gopkg.in/urfave/cli.v1" // imports as package "cli"
)

const (
	// Default host value borrowed from github.com/docker/docker/opts
	defaultHost = "unix:///var/run/docker.sock"
)

var (
	host       = defaultHost
	tlsOptions = tlsconfig.Options{}
)

func main() {
	app := cli.NewApp()
	app.Name = "swarm-prune"
	app.Version = "v0.1"
	app.Compiled = time.Now()
	app.Usage = "swarm wide prune command"
	app.Authors = []cli.Author{
		cli.Author{
			Name:  "Ken Cochrane",
			Email: "@KenCochrane",
		},
	}

	app.Flags = []cli.Flag{
		cli.StringFlag{
			Name:        "host, H",
			Value:       defaultHost,
			Usage:       "Docker Swarm manager host url",
			Destination: &host,
		},
		cli.StringFlag{
			Name:        "tlscacert",
			Value:       "",
			Usage:       "TLS CA cert",
			Destination: &tlsOptions.CAFile,
		},
		cli.StringFlag{
			Name:        "tlscert",
			Value:       "",
			Usage:       "TLS cert",
			Destination: &tlsOptions.CertFile,
		},
		cli.StringFlag{
			Name:        "tlskey",
			Value:       "",
			Usage:       "TLS key",
			Destination: &tlsOptions.KeyFile,
		},
		cli.BoolTFlag{
			Name:        "tlsverify",
			Usage:       "True to skip TLS",
			Destination: &tlsOptions.InsecureSkipVerify,
		},
	}

	app.Commands = []cli.Command{
		{
			Name:        "system",
			Usage:       "This will remove: all stopped containers, all volumes not used by at least one container, all dangling images, all unused networks",
			Description: "WARNING! This will remove all stopped containers, orphaned in your swarm",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "force, F", Usage: "Do not prompt for confirmation"},
				cli.BoolFlag{Name: "all, A", Usage: "This will remove all images without at least one container associated to them"},
			},
			Action: func(c *cli.Context) error {
				validate(host)

				ok := verifyOK(c.Bool("force"))
				if !ok {
					return cli.NewExitError("Ok, will not do anything. exiting now.", 0)
				}
				var spaceReclaimed uint64
				var totalSpaceReclaimed uint64
				nodes := SwarmNodes(host)
				for _, node := range nodes {
					spaceReclaimed = 0
					fmt.Println("###### ", node.Description.Hostname, " ", node.Spec.Role)
					nodeAddr := "tcp://" + node.Description.Hostname + ":2375"
					sr, err := containerPrune(nodeAddr)
					if err != nil {
						fmt.Println(err)
					}
					spaceReclaimed += sr

					sr, err = imagePrune(nodeAddr, c.Bool("all"))
					if err != nil {
						fmt.Println(err)
					}
					spaceReclaimed += sr

					err = networkPrune(nodeAddr)
					if err != nil {
						fmt.Println(err)
					}
					spaceReclaimed += sr

					sr, err = volumePrune(nodeAddr)
					if err != nil {
						fmt.Println(err)
					}
					spaceReclaimed += sr

					totalSpaceReclaimed += spaceReclaimed
					fmt.Println("  Node reclaimed space:", units.HumanSize(float64(spaceReclaimed)))
				}
				fmt.Println("\n  Total Swarm reclaimed space:", units.HumanSize(float64(totalSpaceReclaimed)))
				return nil
			},
		},
		{
			Name:        "containers",
			Usage:       "prune containers swarm wide",
			Description: "WARNING! This will remove all stopped containers in your swarm",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "force, F", Usage: "Do not prompt for confirmation"},
			},
			Action: func(c *cli.Context) error {
				validate(host)
				ok := verifyOK(c.Bool("force"))
				if !ok {
					return cli.NewExitError("Ok, will not do anything. exiting now.", 0)
				}
				var totalSpaceReclaimed uint64
				nodes := SwarmNodes(host)
				for _, node := range nodes {
					fmt.Println("###### ", node.Description.Hostname, " ", node.Spec.Role)
					nodeAddr := "tcp://" + node.Description.Hostname + ":2375"
					spaceReclaimed, err := containerPrune(nodeAddr)
					if err != nil {
						fmt.Println(err)
					}
					totalSpaceReclaimed += spaceReclaimed
					fmt.Println("  Node reclaimed space:", units.HumanSize(float64(spaceReclaimed)))
				}
				fmt.Println("\n    Total Swarm reclaimed space:", units.HumanSize(float64(totalSpaceReclaimed)))
				return nil
			},
		},
		{
			Name:        "images",
			Usage:       "prune images swarm wide",
			Description: "This will remove all dangling images.",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "force, F", Usage: "Do not prompt for confirmation"},
				cli.BoolFlag{Name: "all, A", Usage: "This will remove all images without at least one container associated to them"},
			},
			Action: func(c *cli.Context) error {
				validate(host)
				ok := verifyOK(c.Bool("force"))
				if !ok {
					return cli.NewExitError("Ok, will not do anything. exiting now.", 0)
				}
				nodes := SwarmNodes(host)
				var totalSpaceReclaimed uint64
				for _, node := range nodes {
					fmt.Println("###### ", node.Description.Hostname, " ", node.Spec.Role)
					nodeAddr := "tcp://" + node.Description.Hostname + ":2375"
					spaceReclaimed, err := imagePrune(nodeAddr, c.Bool("all"))
					if err != nil {
						fmt.Println(err)
					}
					totalSpaceReclaimed += spaceReclaimed
					fmt.Println("  Node reclaimed space:", units.HumanSize(float64(spaceReclaimed)))
				}
				fmt.Println("\n    Total Swarm reclaimed space:", units.HumanSize(float64(totalSpaceReclaimed)))
				return nil
			},
		},
		{
			Name:        "volumes",
			Usage:       "prune volumes swarm wide.",
			Description: "WARNING: This will remove all volumes not used by at least one container",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "force, F", Usage: "Do not prompt for confirmation"},
			},
			Action: func(c *cli.Context) error {
				validate(host)
				ok := verifyOK(c.Bool("force"))
				if !ok {
					return cli.NewExitError("Ok, will not do anything. exiting now.", 0)
				}
				var totalSpaceReclaimed uint64
				nodes := SwarmNodes(host)
				for _, node := range nodes {
					fmt.Println("###### ", node.Description.Hostname, " ", node.Spec.Role)
					nodeAddr := "tcp://" + node.Description.Hostname + ":2375"
					spaceReclaimed, err := volumePrune(nodeAddr)
					if err != nil {
						fmt.Println(err)
					}
					totalSpaceReclaimed += spaceReclaimed
					fmt.Println("  Node reclaimed space:", units.HumanSize(float64(spaceReclaimed)))
				}
				fmt.Println("\n    Total Swarm reclaimed space:", units.HumanSize(float64(totalSpaceReclaimed)))
				return nil
			},
		},
		{
			Name:        "networks",
			Usage:       "complete a task on the list",
			Description: "WARNING: This will remove all networks not being used",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "force, F", Usage: "Do not prompt for confirmation"},
			},
			Action: func(c *cli.Context) error {
				validate(host)
				ok := verifyOK(c.Bool("force"))
				if !ok {
					return cli.NewExitError("Ok, will not do anything. exiting now.", 0)
				}
				nodes := SwarmNodes(host)
				for _, node := range nodes {
					fmt.Println("###### ", node.Description.Hostname, " ", node.Spec.Role)
					nodeAddr := "tcp://" + node.Description.Hostname + ":2375"
					err := networkPrune(nodeAddr)
					if err != nil {
						fmt.Println(err)
					}
				}
				return nil
			},
		},
		{
			Name:        "df",
			Usage:       "run docker system df on all nodes",
			Description: "This will show disk usage for all nodes.",
			Flags: []cli.Flag{
				cli.BoolFlag{Name: "verbose, V", Usage: "Verbose output"},
			},
			Action: func(c *cli.Context) error {
				validate(host)
				nodes := SwarmNodes(host)
				for _, node := range nodes {
					fmt.Println("########## ", node.Description.Hostname, " ", node.Spec.Role)
					nodeAddr := "tcp://" + node.Description.Hostname + ":2375"
					err := df(nodeAddr, c.Bool("verbose"), c)
					if err != nil {
						fmt.Println(err)
					}
					fmt.Println("#############################################")
				}
				return nil
			},
		},
	}

	app.Run(os.Args)
}
