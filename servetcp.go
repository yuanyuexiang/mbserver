package mbserver

import (
	"errors"
	"io"
	"log"
	"net"
	"strings"
	"time"
)

func (s *Server) accept(listen net.Listener) error {
	for {
		conn, err := listen.Accept()
		if err != nil {
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			log.Printf("Unable to accept connections: %#v\n", err)
			return err
		}

		go func(conn net.Conn) {
			defer conn.Close()

			for {
				packet := make([]byte, 512)
				bytesRead, err := conn.Read(packet)
				if err != nil {
					if err != io.EOF {
						log.Printf("read error %v\n", err)
					}
					return
				}
				// Set the length of the packet to the number of read bytes.
				packet = packet[:bytesRead]

				frame, err := NewTCPFrame(packet)
				if err != nil {
					log.Printf("bad packet error %v\n", err)
					return
				}

				request := &Request{conn, frame}

				s.requestChan <- request
			}
		}(conn)
	}
}

// ListenTCP starts the Modbus server listening on "address:port".
func (s *Server) ListenTCP(addressPort string) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Printf("Failed to Listen: %v\n", err)
		return err
	}
	s.listeners = append(s.listeners, listen)
	go s.accept(listen)
	return err
}

func (s *Server) acceptX(listen net.Listener, report func(address string, err error)) error {
	for {
		conn, err := listen.Accept()
		if err != nil {
			if conn != nil {
				report(conn.RemoteAddr().String(), err)
			} else {
				report("0.0.0.0:0", err)
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil
			}
			log.Printf("Unable to accept connections: %#v\n", err)
			return err
		}
		report(conn.RemoteAddr().String(), nil)
		go func(conn net.Conn) {
			defer conn.Close()
			defer report(conn.RemoteAddr().String(), errors.New("connection close"))
			for {
				packet := make([]byte, 512)
				conn.SetReadDeadline(time.Now().Add(5 * time.Second))
				bytesRead, err := conn.Read(packet)
				if err != nil {
					if err != io.EOF {
						log.Printf("read error %v\n", err)
					}
					return
				}
				// Set the length of the packet to the number of read bytes.
				packet = packet[:bytesRead]

				frame, err := NewTCPFrame(packet)
				if err != nil {
					log.Printf("bad packet error %v\n", err)
					report(conn.RemoteAddr().String(), err)
					return
				}

				request := &Request{conn, frame}

				s.requestChan <- request
			}
		}(conn)
	}
}

// ListenTCP starts the Modbus server listening on "address:port".
func (s *Server) ListenTCPX(addressPort string, report func(address string, err error)) (err error) {
	listen, err := net.Listen("tcp", addressPort)
	if err != nil {
		log.Printf("Failed to Listen: %v\n", err)
		report("0.0.0.0:0", err)
		return err
	}
	s.listeners = append(s.listeners, listen)
	go s.acceptX(listen, report)
	return err
}
