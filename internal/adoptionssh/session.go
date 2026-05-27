package adoptionssh

import (
	"errors"
	"io"
	"log"
	"net"

	"golang.org/x/crypto/ssh"
)

// serve accepts SSH connections until the listener closes. Each accepted
// connection is handled independently so slow adoption clients do not block new
// attempts.
func (s *Server) serve(config *ssh.ServerConfig) {
	for {
		conn, err := s.listener.Accept()
		if err != nil {
			if errors.Is(err, net.ErrClosed) {
				return
			}
			log.Printf("adoption ssh accept failed: %v", err)
			continue
		}
		go s.handleConn(conn, config)
	}
}

// handleConn completes SSH setup and accepts only session channels, matching
// the narrow command surface required by UniFi advanced adoption.
func (s *Server) handleConn(conn net.Conn, config *ssh.ServerConfig) {
	sshConn, chans, reqs, err := ssh.NewServerConn(conn, config)
	if err != nil {
		_ = conn.Close()
		return
	}
	log.Printf("adoption ssh login user=%s remote=%s", sshConn.User(), sshConn.RemoteAddr())
	go ssh.DiscardRequests(reqs)

	for newChannel := range chans {
		if newChannel.ChannelType() != "session" {
			_ = newChannel.Reject(ssh.UnknownChannelType, "only session channels are supported")
			continue
		}
		channel, requests, err := newChannel.Accept()
		if err != nil {
			continue
		}
		go s.handleSession(channel, requests)
	}
}

// handleSession supports exec and a tiny interactive shell, routing all command
// text through Handler instead of a host shell.
func (s *Server) handleSession(channel ssh.Channel, requests <-chan *ssh.Request) {
	defer func() {
		_ = channel.Close()
	}()
	for req := range requests {
		switch req.Type {
		case "pty-req", "env":
			_ = req.Reply(true, nil)
		case "exec":
			var payload struct {
				Command string
			}
			if err := ssh.Unmarshal(req.Payload, &payload); err != nil {
				_ = req.Reply(false, nil)
				return
			}
			_ = req.Reply(true, nil)
			output, status := s.handler.Execute(payload.Command)
			_, _ = io.WriteString(channel, output)
			sendExitStatus(channel, status)
			return
		case "shell":
			_ = req.Reply(true, nil)
			s.handler.Shell(channel)
			sendExitStatus(channel, 0)
			return
		default:
			_ = req.Reply(false, nil)
		}
	}
}

// sendExitStatus mirrors normal SSH command completion so controller clients do
// not need special-case behavior for the shim.
func sendExitStatus(channel ssh.Channel, status int) {
	_, _ = channel.SendRequest("exit-status", false, ssh.Marshal(struct {
		Status uint32
	}{Status: uint32(status)}))
}
