package mbserver

import (
	"io"
	"log"

	"encoding/hex"
	"strconv"

	"github.com/goburrow/serial"
)

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTU(serialConfig *serial.Config) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Printf("failed to open %s: %v\n", serialConfig.Address, err)
		return
	}
	s.ports = append(s.ports, port)
	go s.acceptSerialRequests(port)
	return err
}

func (s *Server) acceptSerialRequests(port serial.Port) {
SkipFrameError:
	for {
		buffer := make([]byte, 512)

		bytesRead, err := port.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			return
		}
		if bytesRead != 0 {

			// Set the length of the packet to the number of read bytes.
			packet := buffer[:bytesRead]

			frame, err := NewRTUFrame(packet)
			if err != nil {
				log.Printf("bad serial frame error %v\n", err)
				//The next line prevents RTU server from exiting when it receives a bad frame. Simply discard the erroneous
				//frame and wait for next frame by jumping back to the beginning of the 'for' loop.
				log.Printf("Keep the RTU server running!!\n")
				continue SkipFrameError
				//return
			}

			request := &Request{port, frame}

			s.requestChan <- request
		}
	}
}

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTUForLinux(serialConfig *serial.Config, report func(err error)) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Printf("failed to open %s: %v\n", serialConfig.Address, err)
		return
	}
	s.ports = append(s.ports, port)
	go s.acceptSerialRequestsForLinux(port, report)
	return err
}

func (s *Server) acceptSerialRequestsForLinux(port serial.Port, report func(err error)) {
SkipFrameError:
	for {
		buffer := make([]byte, 512)

		bytesRead, err := port.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			report(err)
			return
		}
		log.Println("收到"+strconv.Itoa(bytesRead)+"字节:", hex.EncodeToString(buffer[:bytesRead]))

		if bytesRead != 0 {

			// Set the length of the packet to the number of read bytes.
			packet := buffer[:bytesRead]

			frame, err := NewRTUFrame(packet)
			if err != nil {
				log.Printf("bad serial frame error %v\n", err)
				//The next line prevents RTU server from exiting when it receives a bad frame. Simply discard the erroneous
				//frame and wait for next frame by jumping back to the beginning of the 'for' loop.
				log.Printf("Keep the RTU server running!!\n")
				continue SkipFrameError
				//return
			}

			request := &Request{port, frame}

			s.requestChan <- request
		}
	}
}

// ListenRTU starts the Modbus server listening to a serial device.
// For example:  err := s.ListenRTU(&serial.Config{Address: "/dev/ttyUSB0"})
func (s *Server) ListenRTUForWindows(serialConfig *serial.Config, report func(err error)) (err error) {
	port, err := serial.Open(serialConfig)
	if err != nil {
		log.Printf("failed to open %s: %v\n", serialConfig.Address, err)
		return
	}
	s.ports = append(s.ports, port)
	go s.acceptSerialRequestsForWindows(port, report)
	return err
}

var (
	data  = make([]byte, 512)
	total = 0
)

func (s *Server) acceptSerialRequestsForWindows(port serial.Port, report func(err error)) {
SkipFrameError:
	for {
		buffer := make([]byte, 512)

		bytesRead, err := port.Read(buffer)
		if err != nil {
			if err != io.EOF {
				log.Printf("serial read error %v\n", err)
			}
			report(err)
			return
		}
		log.Println("收到"+strconv.Itoa(bytesRead)+"字节:", hex.EncodeToString(buffer[:bytesRead]))
		if bytesRead <= 2 {
			data = buffer[:bytesRead]
			total = bytesRead
			continue SkipFrameError
		}
		buffer = append(data[:total], buffer...)
		bytesRead = bytesRead + total
		data = []byte{}
		total = 0
		if bytesRead >= 5 {

			// Set the length of the packet to the number of read bytes.
			packet := buffer[:bytesRead]

			frame, err := NewRTUFrame(packet)
			if err != nil {
				log.Printf("bad serial frame error %v\n", err)
				//The next line prevents RTU server from exiting when it receives a bad frame. Simply discard the erroneous
				//frame and wait for next frame by jumping back to the beginning of the 'for' loop.
				log.Printf("Keep the RTU server running!!\n")
				continue SkipFrameError
				//return
			}

			request := &Request{port, frame}

			s.requestChan <- request
		}
	}
}
