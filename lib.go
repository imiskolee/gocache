package gocache

import (
	"encoding/json"
	cache2 "github.com/astaxie/beego/cache"
	"github.com/ngaut/log"
	"net"
	"time"
)

const (
	MagicNumber 	= 0x00abcdef
	CommandPut 		= 0x1
	CommandDelete 	= 0x2
	)

var Pool *pool = nil

type Opt struct {
	BroadcastAddr string
	BroadcastPort int
}

type packet struct {
	MagicNumber int32 	`json:"g"`
	CacheKey string 	`json:"k"`
	Command uint8 		`json:"c"`
}

type pool struct {
	opt *Opt
	cache cache2.Cache
	localIps []string
}

func Init(opt *Opt) {
	Pool = &pool{
		opt:   opt,
		cache: cache2.NewMemoryCache(),
	}
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		panic(err)
	}
	for _,addr := range addrs {
		r,_,_ := net.ParseCIDR(addr.String())
		Pool.localIps = append(Pool.localIps,r.String())
	}
	go Pool.listenUDP()
}

func (p *pool) Put(key string,val interface{},timeout time.Duration) {
	p.cache.Put(key,val,timeout)
	_ = p.broadcast(key,CommandPut)
}

func (p *pool) Get(key string) (interface{},bool) {
	if !p.cache.IsExist(key) {
		return nil, false
	}
	v := p.cache.Get(key)
	return v,true
}

func (p *pool) Delete(key string) error {
	err :=  p.cache.Delete(key)
	if err == nil {
		_ = p.broadcast(key,CommandDelete)
	}
	return err
}

func (p *pool) listenUDP() {
	server,err := net.ListenUDP("udp",&net.UDPAddr{IP:net.IPv4zero,Port:p.opt.BroadcastPort})
	if err != nil {
		panic("can't listen udp")
	}

	for {
		start:
		cache := make([]byte,1024)
		n,remoteAddr,err := server.ReadFromUDP(cache)
		if err != nil {
			log.Errorf("[GoCache] can't read data from udp client.")
			continue
		}
		for _,localIP := range p.localIps {
			if localIP == remoteAddr.IP.String() {
				goto start
			}
		}
		log.Debugf("[GoCache] receiving message from %s:%d msg_len=%d",remoteAddr.IP,remoteAddr.Port,n)
		var packet packet
		if err := json.Unmarshal(cache,&packet); err != nil {
			log.Errorf("[GoCache] can't parse data from udp client:" + err.Error())
			continue
		}
		if packet.MagicNumber != MagicNumber {
			log.Errorf("[GoCache] dismatch magic number %d != %d:",packet.MagicNumber,MagicNumber)
			continue
		}
		switch packet.Command {
			case CommandPut,CommandDelete:
				p.Delete(packet.CacheKey)
		}
	}
}


func (p *pool) broadcast(key string,command uint8) error {
	packet := packet{
		MagicNumber: MagicNumber,
		CacheKey:    key,
		Command:     command,
	}
	m,err := json.Marshal(packet)
	if err != nil {
		return err
	}
	src := &net.UDPAddr{IP:net.IPv4zero,Port:0}
	dst := &net.UDPAddr{IP:net.ParseIP(p.opt.BroadcastAddr),Port:p.opt.BroadcastPort}
	conn,err := net.ListenUDP("udp",src)
	if err != nil {
		return err
	}
	defer conn.Close()
	_,err = conn.WriteToUDP(m,dst)
	if err != nil {
		return err
	}
	return nil
}

