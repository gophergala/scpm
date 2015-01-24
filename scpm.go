package scpm

import (
	"errors"
	"golang.org/x/crypto/ssh"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	// "golang.org/x/crypto/ssh/agent"
)

type Identity struct {
	//TODO ssh.Conn
	Keys []ssh.ClientConfig
}

type Host struct {
	User   string
	Addr   string
	Input  string
	Output string
	Port   int
	Identity
}

func NewHost(host string, key string, port int) (h Host, err error) {
	h = Host{}
	if strings.Index(host, "@") == -1 {
		h.User = os.Getenv("LOGNAME")
	} else {
		arrHost := strings.Split(host, "@")
		h.User = arrHost[0]
		host = strings.Replace(host, h.User+"@", "", -1)
	}
	if strings.Index(host, ":") == -1 {
		err = errors.New("host incorrect")
		return h, err
	}
	arrHost := strings.Split(host, ":")
	h.Addr = arrHost[0]
	h.Output = arrHost[1]
	h.Port = port
	keys := []string{
		key,
		os.Getenv("HOME") + "/.ssh/id_rsa",
		os.Getenv("HOME") + "/.ssh/id_dsa",
		os.Getenv("HOME") + "/.ssh/id_ecdsa",
	}
	for _, k := range keys {
		//create ssh.Config from private keys
		f, err := os.Open(k)
		if err != nil {
			continue
		}
		data, err := ioutil.ReadAll(f)
		if err != nil {
			return h, err
		}
		signer, err := ssh.ParsePrivateKey(data)
		if err != nil {
			return h, err
		}
		h.Keys = append(h.Keys, ssh.ClientConfig{
			User: h.User,
			Auth: []ssh.AuthMethod{
				ssh.PublicKeys(signer),
			},
		})
		f.Close()
	}
	return
}

func (h *Host) Copy(file string) {

}

type Scp struct {
	hosts   []Host
	timeout time.Duration
	input   string
	wg      sync.WaitGroup
	done    chan bool
}

func New(hosts []Host, timeout time.Duration, path string) (scp *Scp, err error) {
	absPath, err := filepath.Abs(path)
	if err != nil {
		return
	}
	if len(hosts) == 0 {
		err = errors.New("hosts is nil")
		return
	}
	scp = new(Scp)
	scp.hosts = hosts
	scp.timeout = timeout
	scp.input = absPath
	scp.done = make(chan bool)
	log.Printf("%#v\n", scp)
	return
}

func (s *Scp) Run(quit chan bool) {
	for {
		select {
		case <-quit:
			quit <- true
		case <-s.done:
			quit <- true
		}
	}
}
