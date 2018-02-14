package iptbutil

import (
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	config "github.com/ipfs/go-ipfs/repo/config"
)

var _ TestbedNode = (*FilecoinNode)(nil)

type FilecoinNode struct {
	Dir     string
	PeerID  string
	ApiPort string
}

func (fn *FilecoinNode) Init() error {
	return os.MkdirAll(fn.Dir, 0755)
}

func (fn *FilecoinNode) envForDaemon() ([]string, error) {
	envs := os.Environ()
	dir := fn.Dir
	npath := "FIL_PATH=" + dir
	for i, e := range envs {
		p := strings.Split(e, "=")
		if p[0] == "FIL_PATH" {
			envs[i] = npath
			return envs, nil
		}
	}

	return append(envs, npath, "FIL_API="+fn.ApiPort), nil
}

func (fn *FilecoinNode) Start(args []string) error {
	env, err := fn.envForDaemon()
	if err != nil {
		return err
	}

	args = append([]string{"--cmdapiaddr=" + fn.ApiPort, "--swarmlisten=/ip4/127.0.0.1/tcp/0"}, args...)
	if err := startProcess("go-filecoin", "daemon", args, fn.Dir, env); err != nil {
		return err
	}

	// wait for node to come online
	time.Sleep(time.Millisecond * 100)

	for i := 0; i < 10; i++ {
		pid, err := fn.RunCmd("go-filecoin", "id", "--format=<id>")
		if err == nil {
			fn.PeerID = strings.TrimSpace(pid)
			break
		}
		fmt.Printf("get id error: %s, retrying...", err)
		time.Sleep(time.Millisecond * 100)
	}
	fmt.Println("PEER ID IS: ", fn.PeerID)

	return nil
}

func (fn *FilecoinNode) String() string {
	return fn.PeerID
}

func (fn *FilecoinNode) APIAddr() (string, error) {
	panic("NYI")
}

func (fn *FilecoinNode) GetAttr(attr string) (string, error) {
	panic("NYI")
}

func (fn *FilecoinNode) SetAttr(attr, val string) error {
	panic("NYI")
}

func (fn *FilecoinNode) GetConfig() (*config.Config, error) {
	panic("NYI")
}

func (fn *FilecoinNode) WriteConfig(cfg *config.Config) error {
	panic("NYI")
}

func (fn *FilecoinNode) GetPeerID() string {
	return fn.PeerID
}

func (fn *FilecoinNode) Kill() error {
	pid, err := getPID(fn.Dir)
	if err != nil {
		return err
	}

	return killPid(pid, fn.Dir)
}

func (fn *FilecoinNode) RunCmd(args ...string) (string, error) {
	env, err := fn.envForDaemon()
	if err != nil {
		return "", fmt.Errorf("error getting env: %s", err)
	}

	return runCmd(args, env)
}

func (fn *FilecoinNode) Shell() error {
	shell := os.Getenv("SHELL")
	if shell == "" {
		return fmt.Errorf("couldnt find shell!")
	}

	nenvs := []string{"FIL_PATH=" + fn.Dir}

	nodes, err := LoadNodes()
	if err != nil {
		return err
	}

	for i, n := range nodes {
		peerid := n.GetPeerID()
		if peerid == "" {
			return fmt.Errorf("failed to check peerID")
		}

		nenvs = append(nenvs, fmt.Sprintf("NODE%d=%s", i, peerid))
	}
	nenvs = append(os.Environ(), nenvs...)

	return syscall.Exec(shell, []string{shell}, nenvs)
}

func (fn *FilecoinNode) BinName() string {
	return "go-filecoin"
}
