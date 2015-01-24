package scpm

import (
	"errors"
	"os"
	"strings"
)

type Identity struct {
	//TODO ssh.Conn
	Keys []string
}

type Host struct {
	User string
	Addr string
	Path string
	Identity
}

func NewHost(host string, key string) (h Host, err error) {
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
		return
	}
	arrHost := strings.Split(host, ":")
	h.Addr = arrHost[0]
	h.Path = arrHost[1]
	h.Identity.Keys = []string{
		key,
		os.Getenv("HOME") + "/.ssh/id_rsa",
		os.Getenv("HOME") + "/.ssh/id_dsa",
		os.Getenv("HOME") + "/.ssh/id_ecdsa",
	}

	return
}

type ff struct {
}

func New(hosts []Host) {

}
