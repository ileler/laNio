package lanio

import (
    "log"
    "net"
    "errors"
)

type io struct {
    mConn *net.UDPConn
    uConn *net.UDPConn

    config    CG
    handler   IHandler
    sendMsg   chan Msg
}

func NewIO(c CG, h IHandler) (IO, error) {
    if "" == c.MCastAddr {
        return nil, errors.New("MCastAddr is empty")
    }
    if 0 == c.MCastPort {
        return nil, errors.New("MCastPort is empty")
    }
    return &io{
        config:    c,
        handler:   h,
        sendMsg:   make(chan Msg),
    }, nil
}

func (_io *io) Start() error {
    //listen
    addr := &net.UDPAddr{IP: net.ParseIP(_io.config.MCastAddr), Port: _io.config.MCastPort}
    mcon, cerr := net.ListenMulticastUDP("udp4", nil, addr)
    if cerr != nil {
        return cerr
    }
    _io.mConn = mcon

    log.Printf("listening multicast [%s:%d] ...", addr.IP.String(), addr.Port)

    if 0 != _io.config.UCastPort {
        addr = &net.UDPAddr{IP: IP(), Port: _io.config.UCastPort}
        _io.uConn, cerr = net.ListenUDP("udp4", addr)
        if cerr != nil {
            return cerr
        }

        log.Printf("listening udp [%s:%d] ...", addr.IP.String(), addr.Port)

        //unicast read
        go func() {
            defer _io.uConn.Close()

            buf := make([]byte, 2048)
            for {
                size, _addr, err := _io.uConn.ReadFromUDP(buf)
                if err != nil {
                    log.Printf("UNICAST<< closed connection", err)
                    return
                } else if size > 0 {
                    log.Printf("UNICAST<< recv msg from [%s:%d]: %s", _addr.IP.String(), _addr.Port, string(buf[0:size]))

                    _msg := _io.handler.UReply(Msg{ IP: _addr.IP.String(), Port: _addr.Port, Data: buf[0:size] })

                    if _msg != nil {
                        _io.Send(Msg{ IP: _addr.IP.String(), Port: _addr.Port, Data: _msg })
                    }
                }
            }
        }()
    }

    //multicast read
    go func() {
        defer _io.mConn.Close()

        buf := make([]byte, 2048)
        for {
            size, _addr, err := _io.mConn.ReadFromUDP(buf)
            if err != nil {
                log.Printf("MULTICAST<< closed connection", err)
                return
            } else if size > 0 {
                log.Printf("MULTICAST<< recv msg from [%s:%d]: %s", _addr.IP.String(), _addr.Port, string(buf[0:size]))

                _msg := _io.handler.MReply(Msg{ IP: _addr.IP.String(), Port: _addr.Port, Data: buf[0:size] })

                if _msg != nil {
                    _io.Send(Msg{ IP: _addr.IP.String(), Port: _addr.Port, Data: _msg })
                }
            }
        }
    }()

    //write
    go func() {
        for {
            _io._write(<-_io.sendMsg)
        }
    }()

    return nil
}

func (_io *io) Send(msg Msg) error {
    _io.sendMsg <- msg
    return nil
}

func (_io *io) _write(msg Msg) {
    if len(msg.Data) < 1 {
        return
    }
    var casttype string
    if msg.IP == _io.config.MCastAddr {
        casttype = "MULTICAST"
    } else {
        casttype = "UNICAST"
    }
    wsize, err := _io.mConn.WriteToUDP(msg.Data, &net.UDPAddr{ IP: net.ParseIP(msg.IP), Port: msg.Port })
    if err != nil {
        log.Printf(casttype + ">> ERR::", err)
        return
    }
    log.Printf(casttype + ">> send to [%s:%d] - msg:[%s], %d bytes!", msg.IP, msg.Port, string(msg.Data), wsize)
}
