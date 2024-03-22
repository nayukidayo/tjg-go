package rfid

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"log"
	"net"
	"time"

	"github.com/nats-io/nats.go"
	"github.com/nayukidayo/tjg-go/env"
)

type config struct {
	addr   string
	maxBuf int
	nc     *nats.Conn

	delimiter    []byte
	overhead     int
	lengthOffset int
	lengthBytes  int

	code         []byte
	codeOffset   int
	deviceOffset int
	deviceBytes  int
	tagNumOffset int
}

func newConfig(nc *nats.Conn) *config {
	return &config{
		addr:   env.GetStr("RFID_ADDR", ":54329"),
		maxBuf: env.GetInt("RFID_MAX_BUF", 512),
		nc:     nc,

		delimiter:    []byte{0x43, 0x54},
		overhead:     4,
		lengthOffset: 2,
		lengthBytes:  2,

		code:         []byte{0x45},
		codeOffset:   5,
		deviceOffset: 4,
		deviceBytes:  1,
		tagNumOffset: 14,
	}
}

func (c *config) pub(rs *Result) {
	data, err := json.Marshal(rs)
	if err != nil {
		log.Println(err)
		return
	}
	c.nc.Publish("tjg.rfid", data)
}

type parser struct {
	conf   *config
	buf    []byte
	device string
}

func newParser(conf *config) *parser {
	return &parser{
		conf: conf,
		buf:  make([]byte, 0, conf.maxBuf),
	}
}

func (p *parser) transform(chunk []byte) {
	data := append(p.buf, chunk...)
	pos := bytes.Index(data, p.conf.delimiter)
	for pos != -1 {
		if len(data) < pos+p.conf.lengthOffset+p.conf.lengthBytes {
			break
		}
		length := binary.BigEndian.Uint16(data[pos+p.conf.lengthOffset : pos+p.conf.lengthOffset+p.conf.lengthBytes])
		total := pos + int(length) + p.conf.overhead
		if len(data) < total {
			break
		}
		p.filter(data[pos:total])
		data = data[total:]
		pos = bytes.Index(data, p.conf.delimiter)
	}
	if len(data) > p.conf.maxBuf {
		p.buf = make([]byte, 0, p.conf.maxBuf)
	} else {
		p.buf = data
	}
}

func (p *parser) filter(chunk []byte) {
	if bytes.Equal(chunk[p.conf.codeOffset:p.conf.codeOffset+len(p.conf.code)], p.conf.code) && p.check(chunk) {
		if p.device == "" {
			p.device = hex.EncodeToString(chunk[p.conf.deviceOffset : p.conf.deviceOffset+p.conf.deviceBytes])
		}
		rs := new(Result)
		rs.Type = "RFID"
		rs.Device = p.device
		rs.TS = time.Now().UnixMilli()
		rs.Data = p.decode(chunk)
		p.conf.pub(rs)
	}
}

func (p *parser) check(chunk []byte) bool {
	sum := 0
	lth := len(chunk) - 1
	for i := 0; i < lth; i++ {
		sum += int(chunk[i])
	}
	sum = 256 - (sum % 256)
	return sum == int(chunk[lth])
}

func (p *parser) decode(chunk []byte) []RFID {
	num := int(chunk[p.conf.tagNumOffset])
	tags := make([]RFID, 0, num)
	start := p.conf.tagNumOffset + 1
	for i := 0; i < num; i++ {
		end := start + int(chunk[start])
		tag := chunk[start+3 : end]
		tags = append(tags, RFID{
			Tag:  hex.EncodeToString(tag),
			Rssi: int(chunk[end]),
		})
		start = end + 1
	}
	return tags
}

type Result struct {
	Type   string `json:"type"`
	Device string `json:"device"`
	Data   []RFID `json:"data"`
	TS     int64  `json:"ts"`
}

type RFID struct {
	Tag  string `json:"tag"`
	Rssi int    `json:"rssi"`
}

func Server(nc *nats.Conn) {
	conf := newConfig(nc)
	ln, err := net.Listen("tcp", conf.addr)
	if err != nil {
		log.Fatalln(err)
	}
	defer ln.Close()
	log.Println("RFID", ln.Addr().String())
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Fatalln(err)
		}
		go handleConn(conn, conf)
	}
}

func handleConn(conn net.Conn, conf *config) {
	defer conn.Close()
	parser := newParser(conf)
	lth := parser.conf.maxBuf >> 2
	for {
		chunk := make([]byte, lth)
		n, err := conn.Read(chunk)
		if err != nil {
			return
		}
		if n > 0 {
			parser.transform(chunk[:n])
		}
	}
}

// 数据协议
// 4354001c084501c18323121455ae010f010100112233445566778899aabbbe3d
// 4354002c084501c18323121455ae020f010100112233445566778899aabb040f0101e280116060000217299f4b39fa43
//
// 4354 001c 08 45 01 c18323121455ae 01 0f 01 01 00112233445566778899aabb be 3d
//
// 4354 头
// 001c 长度
// 08 地址
// 45 响应码
// 01
// c18323121455ae 设备序列号
// 01 标签总数
// len type ant tag                      rssi
// 0f  01   01  00112233445566778899aabb be
// 3d 校验码
