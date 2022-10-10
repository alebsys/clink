/*
ROADMAP
+ добавить флаги по переключению поиска по разным runtime - contanerd, docker
+ добавить справку
- добавить сопоставление всех интерфейсов на Хосте, флаг -a / --all
+ вынести containerd namespace в отдельную переменную (k8s.io)
*/

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/spf13/pflag"
	"github.com/vishvananda/netlink"
)

const (
	CONTAINERD_TASK_DIR = "/run/containerd/io.containerd.runtime.v2.task"
	CONTAINERD_PIDFILE  = "init.pid"
	CONTAINERD_SOCK_DIR = "/run/containerd/containerd.sock"
)

type container struct {
	ID      string
	PID     int
	Network string
}

var (
	containerID string
	runtime     string
	namespace   string
)

func init() {
	pflag.StringVarP(&containerID, "container.id", "i", "", "Container ID")
	pflag.StringVarP(&runtime, "runtime", "r", "containerd", "Used runtime")
	pflag.StringVarP(&namespace, "namespace", "n", "k8s.io", "Used namespace")
}

// TODO вынести из main() по максимум во внешние функции
func main() {
	pflag.Parse()

	// Check that container id is set
	if containerID == "" {
		fmt.Println("Container ID not set.")
		pflag.Usage()
		os.Exit(1)
	}

	client, err := containerd.New(CONTAINERD_SOCK_DIR)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), namespace)
	containers, err := client.Containers(ctx)
	if err != nil {
		log.Fatal(err)
	}

	isExist := checkExistContainerID(containers, containerID)
	if !isExist {
		log.Fatal(errNotFound)
	}

	containerPID, err := findContainerPid(containerID, namespace)
	if err != nil {
		log.Fatal(err)
	}

	containerNetNsID, err := netlink.GetNetNsIdByPid(containerPID)
	if err != nil {
		log.Fatal(err)
	}

	links, err := netlink.LinkList()
	if err != nil {
		log.Fatal(err)
	}

	containerLink, err := findContainerLink(links, containerNetNsID)
	if err != nil && err != errUsedHostInterface {
		log.Fatal(err)
	}

	container := container{
		containerID,
		containerPID,
		containerLink,
	}

	writeTable(container)
}

// writeTable writes table with data
// TODO внедрить добавление новых строк в цикле при будущем флаге --all
func writeTable(cont container) {
	t := table.NewWriter()
	t.SetOutputMirror(os.Stdout)
	// TODO добавить имя контейнера/Пода
	t.AppendHeader(table.Row{"ID", "PID", "Network Interface"})
	t.AppendRows([]table.Row{
		{cont.ID, cont.PID, cont.Network},
	})
	t.AppendSeparator()
	t.Render()
}

// checkExistContainerID checks if exist containerID in containerd NS, e.g. CONTAINERD_K8S_NAMESPACE
func checkExistContainerID(containers []containerd.Container, containerID string) bool {
	for _, c := range containers {
		if containerID == c.ID() {
			return true
		}
	}

	return false
}

// findContainerPid finds container's PID by container's ID
func findContainerPid(containerID, namespace string) (int, error) {
	path := filepath.Join(CONTAINERD_TASK_DIR, namespace, containerID, CONTAINERD_PIDFILE)

	data, err := os.ReadFile(path)
	if err != nil {
		return 0, errPIDFileNotFound
	}

	pid, err := strconv.Atoi(string(data))
	if err != nil {
		return 0, err
	}

	return pid, nil
}

// findContainerLink finds container's network interface by network namespace ID
func findContainerLink(ll []netlink.Link, netID int) (string, error) {
	for _, l := range ll {
		if l.Attrs().NetNsID == netID && l.Type() == "veth" {
			return l.Attrs().Name, nil
		}

		// TODO добавить явное отображение названия хостового интерфейса
		if l.Attrs().NetNsID == netID && l.Type() != "veth" {
			return "", errUsedHostInterface
		}
	}

	return "", errNotFound
}
