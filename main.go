package main

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"time"

	"github.com/go-irc/irc"
	"github.com/sensu-utils/toolbox"
)

// Config represents the expected config.
type Config struct {
	Nick     string `json:"irc_nick"`
	Server   string `json:"irc_server"`
	Password string `json:"irc_password"`
	SSL      bool   `json:"irc_ssl"`
	Channel  string `json:"irc_channel"`
}

var config Config
var event toolbox.Event

func main() {
	toolbox.InitPlugin("irc", &event, &config)

	errChan := make(chan error)

	go func() {
		var err error
		var rawConn net.Conn
		if config.SSL {
			rawConn, err = tls.Dial("tcp", config.Server, nil)
		} else {
			rawConn, err = net.Dial("tcp", config.Server)
		}
		if err != nil {
			errChan <- err
			return
		}
		defer rawConn.Close()

		conn := irc.NewConn(rawConn)
		if config.Password != "" {
			if err := conn.Writef("PASS :%s", config.Password); err != nil {
				log.Println("failed to send PASS to server")
				return
			}
		}
		if err := conn.Writef("NICK :%s", config.Nick); err != nil {
			log.Println("failed to send NICK to server")
			return
		}
		if err := conn.Writef("USER %s 0.0.0.0 0.0.0.0 :%s", "sensu", "sensu"); err != nil {
			log.Println("failed to set IRC user to server")
			return
		}

		var actionString = "\x0301,04ALERT\x03"
		if event.IsResolution() {
			actionString = "\x0300,03RESOLVED\x03"
		}

		for {
			msg, err := conn.ReadMessage()
			if err != nil {
				errChan <- err
				return
			}

			if msg.Command == "PING" {
				reply := msg.Copy()
				reply.Command = "PONG"
				if err := conn.WriteMessage(reply); err != nil {
					log.Println("failed reply to server PING")
					return
				}
			} else if msg.Command == "001" {
				if err := conn.Writef("JOIN :%s", config.Channel); err != nil {
					log.Printf("failed JOIN IRC Room: %s", config.Channel)
					return
				}
				if err := conn.Writef("PRIVMSG %s :%s %s/%s: %s", config.Channel, actionString, event.Entity.System.Hostname, event.Check.Name, event.Check.Output); err != nil {
					log.Printf("failed message IRC Room %s with message: %s %s/%s: %s", config.Channel, actionString, event.Entity.System.Hostname, event.Check.Name, event.Check.Output)
					return
				}
				if err := conn.Writef("QUIT :bye"); err != nil {
					log.Println("failed to quit IRC server")
					return
				}

				errChan <- nil
				return
			}
		}
	}()

	select {
	case err := <-errChan:
		if err != nil {
			log.Println(err)
			os.Exit(1)
		} else {
			log.Println("Sent message")
		}
	case <-time.After(10 * time.Second):
		log.Fatalln("Message timeout")
	}
}
