package logstash

import "net"

var LOGSTASH *LogstashModel

type LogstashModel struct {
	Network string
	Addr    string
}

func (l *LogstashModel) Open() (net.Conn, error) {
	var err error

	conn, err := net.Dial(l.Network, l.Addr)
	if err != nil {
		return nil, err
	}

	return conn, nil

}
