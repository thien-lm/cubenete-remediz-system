package vcd

import (
    "fmt"
    "net/url"
    "github.com/vmware/go-vcloud-director/v2/govcd"
)

type Config struct {
    User     string
    Password string
    Org      string
    Href     string
    VDC      string
    Insecure bool
}

func (c *Config) NewClient() (*govcd.VCDClient, error) {
    u, err := url.ParseRequestURI(c.Href)
    if err != nil {
        return nil, fmt.Errorf("unable to pass url: %s", err)
    }
    vcdclient := govcd.NewVCDClient(*u, true)
    err = vcdclient.Authenticate(c.User, c.Password, c.Org)
    if err != nil {
        return nil, fmt.Errorf("unable to authenticate: %s", err)
    }
    return vcdclient, nil
}
