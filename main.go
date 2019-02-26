package main

import (
	"crypto/tls"
	"log"
	"net"
	"os"
	"time"

	utils "github.com/belak/sensu-go-utils"
	"github.com/go-irc/irc"
)

type Config struct {
	Nick     string `json:"irc_nick"`
	Server   string `json:"irc_server"`
	Password string `json:"irc_password"`
	SSL      bool   `json:"irc_ssl"`
	Channel  string `json:"irc_channel"`
}

var config Config
var event utils.Event

func main() {
	utils.InitPlugin("irc", &event, &config)

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
			conn.Writef("PASS :%s", config.Password)
		}
		conn.Writef("NICK :%s", config.Nick)
		conn.Writef("USER %s 0.0.0.0 0.0.0.0 :%s", "sensu", "sensu")

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
				conn.WriteMessage(reply)
			} else if msg.Command == "001" {
				conn.Writef("JOIN :%s", config.Channel)
				conn.Writef("PRIVMSG %s :%s %s/%s: %s", config.Channel, actionString, event.Entity.System.Hostname, event.Check.Name, event.Check.Output)
				conn.Writef("QUIT :bye")

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
		log.Println("Message timeout")
	}
}
