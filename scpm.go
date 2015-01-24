package scpm

import (
	// "bytes"
	"errors"
	"fmt"
	"github.com/cheggaaa/pb"
	"golang.org/x/crypto/ssh"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
	// "golang.org/x/crypto/ssh/agent"
)

type Host struct {
	User     string
	Addr     string
	Input    string
	Output   string
	Client   *ssh.Client
	sess     *ssh.Session
	Identity *ssh.ClientConfig
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
	h.Addr = arrHost[0] + ":" + fmt.Sprint(port)
	h.Output = arrHost[1]
	keys := []string{
		key,
		os.Getenv("HOME") + "/.ssh/id_rsa",
		os.Getenv("HOME") + "/.ssh/id_dsa",
		os.Getenv("HOME") + "/.ssh/id_ecdsa",
	}
	h.Identity = &ssh.ClientConfig{User: h.User}
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
		h.Identity.Auth = append(h.Identity.Auth, ssh.PublicKeys(signer))
		f.Close()
	}
	return
}

func (h *Host) Auth() error {
	var err error
	if len(h.Identity.Auth) == 0 {
		return errors.New("TODO dialog password")
	} else {
		h.Client, err = ssh.Dial("tcp", h.Addr, h.Identity)
		if err != nil {
			return err
		}
	}
	return nil
}

func (h *Host) Copy(file string, wg *sync.WaitGroup) {
	h.Input = file
	defer func() {
		wg.Done()
	}()
	if err := h.Auth(); err != nil {
		log.Println("h.Auth", err)
		return
	}
	if err := h.cp(h.Output); err != nil {
		log.Println("h.cp", err)
	}
	// if err := h.sess.Wait(); err != nil {
	// 	log.Println("sess.wait", err)
	// }
}

func (h *Host) exec(cmd string) error {
	var err error
	h.sess, err = h.Client.NewSession()
	if err != nil {
		return err
	}
	defer h.sess.Close()
	return h.sess.Run(cmd)
}

const (
	cmdStat  string = "stat %s"
	cmdCat   string = "cat > %s"
	cmdMkDir string = "mkdir -p %s"
)

//remote cp
func (h *Host) cp(path string) error {
	var err error
	//create remote dir
	dir := filepath.Dir(path)
	if err := h.exec(fmt.Sprintf(cmdStat, dir)); err != nil {
		return h.mkdir(dir)
	}
	//open fd file
	f, err := os.Open(h.Input)
	if err != nil {
		return err
	}
	defer f.Close()
	//open ssh session
	h.sess, err = h.Client.NewSession()
	if err != nil {
		return err
	}
	defer h.sess.Close()
	dest, err := h.sess.StdinPipe()
	if err != nil {
		return err
	}
	if err = h.sess.Start(fmt.Sprintf(cmdCat, path)); err != nil {
		return err
	}

	info, err := f.Stat()
	if err != nil {
		return err
	}
	// create bar
	bar := pb.New(int(info.Size())).SetUnits(pb.U_BYTES).SetRefreshRate(time.Millisecond * 10)
	bar.ShowSpeed = true
	bar.Start()
	defer bar.Finish()

	writer := io.MultiWriter(dest, bar)

	_, err = io.Copy(writer, f)
	return err
}

//remote mkdir
func (h *Host) mkdir(dir string) error {
	return h.exec(fmt.Sprintf(cmdMkDir, dir))
}

type Scp struct {
	hosts   []Host
	timeout time.Duration
	input   string
	wg      *sync.WaitGroup
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
	scp.wg = new(sync.WaitGroup)
	scp.done = make(chan bool)
	return
}

func (s *Scp) Run(quit chan bool) {
	for _, host := range s.hosts {
		s.wg.Add(1)
		go host.Copy(s.input, s.wg)
	}
	go func() {
		s.wg.Wait()
		s.done <- true
	}()
	for {
		select {
		case <-quit:
			quit <- true
		case <-s.done:
			quit <- true
		}
	}
}
