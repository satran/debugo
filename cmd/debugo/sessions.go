package main

import (
	"encoding/json"
	"log"
)

type Session struct {
	connections    []*Conn        // List of all the active connections for this session

	input  chan []byte      //Channel that listens to inputs from various clients
	output chan interface{} // a channel where commands push result into
}

func NewSession() *Session {
	s := Session{}
	s.connections = []*Conn{}

	s.input = make(chan []byte)
	s.output = make(chan interface{})

	return &s
}

func (s *Session) Listen() {
	go s.listenInput()
	go s.listenOutput()
}

func (s *Session) listenInput() {
	for {
		message := <-s.input
		cmd := Command{}
		if err := json.Unmarshal(message, &cmd); err != nil {
			log.Println("Can't decode", err, message)
			continue
		}
		go cmd.Exec(s)
	}
}

func (s *Session) listenOutput() {
	for {
		response := <-s.output
		for _, conn := range s.connections {
			conn.Write(response)
		}
	}
}

func (s *Session) AddConn(c *Conn) {
	s.connections = append(s.connections, c)
}
