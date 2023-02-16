package main

import (
	"encoding/json"
	"fmt"
	"github.com/sirupsen/logrus"
	"github.com/valyala/fasthttp"
	"github.com/valyala/fastjson"
	"math/rand"
	"net"
	"strconv"
	"strings"
	"time"
	"v2board-saver/common"
)

type Saver struct {
	credential string
}

type node struct {
	// Rules/ruleSettings/dnsSettings/Sort didn't find any implementation
	Id              int            `json:"id"`
	GroupId         []string       `json:"group_id"`
	RouteId         []string       `json:"route_id"`
	Name            string         `json:"name"`
	ParentId        *int           `json:"parent_id"`
	Host            string         `json:"host"`
	Port            string         `json:"port"`
	ServerPort      int            `json:"server_port"`
	TLS             int            `json:"tls"`
	Tags            []string       `json:"tags"`
	Rate            string         `json:"rate"`
	Network         string         `json:"network"`
	Rules           *string        `json:"rules"`
	NetworkSettings *network       `json:"networkSettings"`
	RuleSettings    *string        `json:"ruleSettings"`
	DNSSettings     *string        `json:"dnsSettings"`
	Sort            *string        `json:"sort"`
	TLSSettings     map[string]any `json:"tlsSettings"`
	Show            int            `json:"show"`
	CreatedAt       int            `json:"created_at"`
	UpdatedAt       int            `json:"updated_at"`
	Type            string         `json:"type"`
	Online          string         `json:"online"`
	LastCheckAt     string         `json:"last_check_at"`
	LastPushAt      string         `json:"last_push_at"`
	AvailableStatus int            `json:"available_status"`
}

type network struct {
	Path    string            `json:"path"`
	Headers map[string]string `json:"headers"`
}

func NewUpdater() *Saver {
	return &Saver{}
}

func (s *Saver) Start() {
	logrus.Info("starting saver")
	for {
		time.Sleep(time.Duration(common.Config.Interval) * time.Second)
		logrus.Info("try to save")

		if len(s.credential) == 0 {
			logrus.Warn("empty auth data, try login")
			s.login()
		}
		oldNodes := s.getNodes()
		deadNodes := s.check(oldNodes["data"])
		s.save(deadNodes)
	}
}

func (s *Saver) login() {
	post := []byte(fmt.Sprintf("email=%s&password=%s", common.Config.Email, common.Config.Password))
	body, err := common.DoHttp(
		fmt.Sprintf("%s/api/v1/passport/auth/login", common.Config.URL),
		fasthttp.MethodPost,
		map[string]string{
			"accept-encoding": "br",
			"content-type":    "application/x-www-form-urlencoded",
			"content-length":  strconv.Itoa(len(post)),
			"host":            strings.Split(common.Config.URL, "://")[1],
		},
		post,
	)
	if err != nil {
		logrus.WithError(err).Error("failed to login")
		return
	}

	jd, err := fastjson.ParseBytes(body)
	if err != nil {
		logrus.WithError(err).Error("failed to parse login json bytes")
		return
	}
	if jd.GetInt("data", "is_admin") != 1 {
		logrus.WithError(err).Error("you're not admin, please check your config file")
		return
	}
	s.credential = string(jd.GetStringBytes("data", "auth_data"))
	logrus.Info("login success")
}

func (s *Saver) getNodes() map[string][]*node {
	if len(s.credential) == 0 {
		return nil
	}
	logrus.Info("getting nodes")
	body, err := common.DoHttp(
		fmt.Sprintf("%s/api/v1/jobber/server/manage/getNodes", common.Config.URL),
		fasthttp.MethodGet,
		map[string]string{
			"accept-encoding": "br",
			"host":            strings.Split(common.Config.URL, "://")[1],
			"authorization":   s.credential,
		},
		nil,
	)
	if err != nil {
		logrus.WithError(err).Error("failed to parse get nodes json bytes")
		return nil
	}

	nodes := map[string][]*node{
		"data": make([]*node, 0),
	}
	if err := json.Unmarshal(body, &nodes); err != nil {
		logrus.WithError(err).Error("failed to unmarshal nodes json to struct")
		return nil
	}
	return nodes
}

func (s *Saver) check(nodes []*node) []*node {
	if len(nodes) == 0 {
		return nil
	}
	logrus.Info("checking nodes")
	dead := make([]*node, 0)
	for _, node := range nodes {
		conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", node.Host, node.ServerPort), time.Duration(common.Config.Timeout)*time.Second)
		if err != nil {
			logrus.WithError(err).WithFields(logrus.Fields{
				"host": node.Host,
				"port": node.Port,
				"name": node.Name,
			}).Error("node is dead to you")
			dead = append(dead, node)
		} else {
			logrus.WithFields(logrus.Fields{
				"host": node.Host,
				"port": node.Port,
				"name": node.Name,
			}).Info("node is alive to you")
			_ = conn.Close()
		}
	}
	return dead
}

func (s *Saver) save(nodes []*node) {
	if len(nodes) == 0 {
		return
	}
	logrus.Info("saving nodes")
	for _, node := range nodes {
		for {
			node.ServerPort = 1000 + rand.Intn(64536)
			if s.checkPort(node.ServerPort, node.Host) {
				break
			}
		}
		node.Port = strconv.Itoa(node.ServerPort)
		node.UpdatedAt = int(time.Now().Unix())

		form, err := common.StructToForm(node)
		if err != nil {
			logrus.WithError(err).Error("failed to parse struct to form")
			continue
		}
		body, err := common.DoHttp(
			fmt.Sprintf("%s/api/v1/jobber/server/v2ray/save", common.Config.URL),
			fasthttp.MethodPost,
			map[string]string{
				"accept-encoding": "br",
				"host":            strings.Split(common.Config.URL, "://")[1],
				"authorization":   s.credential,
				"content-type":    "application/x-www-form-urlencoded",
				"content-length":  strconv.Itoa(len(form)),
			},
			[]byte(form),
		)
		if err != nil {
			logrus.WithError(err).Error("failed to update node")
			continue
		}
		if string(body) != `{"data":true}` {
			logrus.WithField("body", string(body)).Error("failed to update node")
			continue
		} else {
			logrus.Info("node saved, please update your config file")
		}
	}
}

func (s *Saver) checkPort(port int, host string) bool {
	conn, err := net.DialTimeout("tcp", fmt.Sprintf("%s:%d", host, port), time.Duration(common.Config.Timeout)*time.Second)
	if err == nil {
		_ = conn.Close()
		logrus.WithField("port", port).Warn("port already in use")
		return false
	} else {
		return true
	}
}
