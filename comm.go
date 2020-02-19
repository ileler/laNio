package lanio

import (
    "net"
    "strings"
)

type CG struct {
    MCastAddr string
    MCastPort int
    UCastPort int
}

type Msg struct {
    IP   string
    Port int
    Data []byte
}

type IO interface {
    Start() error
    Send(Msg) error
}

type IHandler interface {
    MReply(msg Msg) []byte
    UReply(msg Msg) []byte
}

// Get preferred outbound ip of this machine
func IP() net.IP {
    conn, err := net.Dial("udp", "8.8.8.8:80")
    if err != nil {
        panic(err)
    }
    defer conn.Close()

    localAddr := conn.LocalAddr().(*net.UDPAddr)

    return localAddr.IP
}

func MAC() net.HardwareAddr {
    interfaces, _ := net.Interfaces()
    for _, interf := range interfaces {
        if addrs, err := interf.Addrs(); err == nil {
            for _, addr := range addrs {
                //log.Println("[", iindex, "-", aindex, "]", interf.Name, ">", addr)

                // only interested in the name with current IP address
                if strings.Contains(addr.String(), IP().String()) {
                    netInterface, err := net.InterfaceByName(interf.Name)

                    if err != nil {
                        return nil
                    }

                    macAddress := netInterface.HardwareAddr

                    // verify if the MAC address can be parsed properly
                    hwAddr, err := net.ParseMAC(macAddress.String())

                    if err == nil {
                        return hwAddr
                    }
                }
            }
        }
    }
    return nil
}
