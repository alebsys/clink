/*
ROADMAP
- добавить флаги по переключению поиска по разным runtime - contanerd, docker
- добавить справку
- добавить сопоставление всех интерфейсов на Хосте, флаг -a / --all
+ вынести containerd namespace в отдельную переменную (k8s.io)

*/

package main

import (
	"context"
	"errors"
	"log"
	"os"
	"strconv"

	"github.com/containerd/containerd"
	"github.com/containerd/containerd/namespaces"
	"github.com/jedib0t/go-pretty/v6/table"
	"github.com/vishvananda/netlink"
)

const (
	CONTAINERD_TASK_DIR      = "/run/containerd/io.containerd.runtime.v2.task"
	CONTAINERD_K8S_NAMESPACE = "k8s.io"
	CONTAINERD_PIDFILE       = "init.pid"
	CONTAINERD_SOCK_DIR      = "/run/containerd/containerd.sock"
)

type container struct {
	ID      string
	PID     int
	Network string
}

// TODO вынести из main() по максимум во внешние функции
func main() {

	if len(os.Args) != 2 {
		log.Fatal("enter only container ID")
	}

	containerID := os.Args[1]

	client, err := containerd.New(CONTAINERD_SOCK_DIR)
	if err != nil {
		log.Fatal(err)
	}
	defer client.Close()

	ctx := namespaces.WithNamespace(context.Background(), CONTAINERD_K8S_NAMESPACE)
	containers, err := client.Containers(ctx)
	if err != nil {
		log.Fatal(err)
	}

	isExist := checkExistContainerID(containers, containerID)
	if !isExist {
		log.Fatal("enter valid container ID")
	}

	containerPID, err := findContainerPid(containerID)
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
	if err != nil {
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
func findContainerPid(containerID string) (int, error) {
	// TODO можно ли найти более лаконичный способ склейки пути
	path := CONTAINERD_TASK_DIR + "/" + CONTAINERD_K8S_NAMESPACE + "/" + containerID + "/" + CONTAINERD_PIDFILE

	pid, err := os.ReadFile(path)
	if err != nil {
		return -1, errors.New("enter wrong container ID")
	}

	sPid := string(pid)
	// TODO можно ли найти более лаконичный способ перевода в int
	nPid, err := strconv.Atoi(sPid)
	if err != nil {
		return -1, err
	}
	return nPid, nil
}

// findContainerLink finds container's network interface by network namespace ID
func findContainerLink(ll []netlink.Link, netID int) (string, error) {
	for _, l := range ll {

		if l.Attrs().NetNsID == netID && l.Type() == "veth" {
			return l.Attrs().Name, nil
		}

		// TODO добавить явное отображение названия хостового интерфейса
		if l.Attrs().NetNsID == netID && l.Type() != "veth" {
			return "Container used Host interface", nil
		}

	}

	return "", errors.New("no matches found")
}
