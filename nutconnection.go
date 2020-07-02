package main

import (
	"bufio"
	"errors"
	"fmt"
	"github.com/go-kit/kit/log/level"
	"net"
	"net/textproto"
	"strings"
)

// protocol details https://networkupstools.org/docs/developer-guide.chunked/ar01s09.html

type connection struct {
	Host       string
	User, Pass string
	UPSName    string
	TCPConn    net.Conn
}

func newConnection(host, user, pass, upsName string) *connection {
	return &connection{host, user, pass, upsName, nil}
}

func (conn *connection) open() error {
	if conn.TCPConn != nil {
		_ = conn.TCPConn.Close()
	}
	dialedConn, err := net.Dial("tcp", conn.Host)
	if err != nil {
		_ = level.Error(logger).Log("msg", "problem connect to NUT server ["+conn.Host+"]", "error", err, "host", conn.Host)
		return err
	}
	conn.TCPConn = dialedConn
	_, err = conn.commandExpect("USERNAME "+conn.User, "OK")
	if err != nil {
		_ = level.Error(logger).Log("msg", err, "ups", conn.UPSName)
		return err
	}
	_, err = conn.commandExpect("PASSWORD "+conn.Pass, "OK")
	if err != nil {
		_ = level.Error(logger).Log("msg", err, "ups", conn.UPSName)
		return err
	}
	_, err = conn.commandExpect("LOGIN "+conn.UPSName, "OK")
	if err != nil {
		_ = level.Error(logger).Log("msg", err, "ups", conn.UPSName)
		return err
	}
	_ = level.Debug(logger).Log("msg", "success login to NUT server for ups name ["+conn.UPSName+"]", "ups", conn.UPSName)
	return nil
}

func (conn *connection) close() {
	if conn.TCPConn != nil {
		_, _ = conn.commandExpect("LOGOUT", "OK")
		_ = conn.TCPConn.Close()
	}
	conn.TCPConn = nil
}

func (conn *connection) command(input string) (string, error) {
	_, _ = fmt.Fprintf(conn.TCPConn, "%s\r\n", input)
	output, err := bufio.NewReader(conn.TCPConn).ReadString('\n')
	if err != nil {
		return "", err
	}
	return strings.TrimSuffix(output, "\n"), nil
}

func (conn *connection) commandList(input string) ([]string, error) {
	_, _ = fmt.Fprintf(conn.TCPConn, "%s\r\n", input)
	reader := bufio.NewReader(conn.TCPConn)
	output := textproto.NewReader(reader)
	var data []string
	for {
		line, err := output.ReadLine()
		if err != nil {
			return data, err
		}
		s := strings.Split(line, " ")
		show := s[2] + ": " + strings.TrimSuffix(strings.TrimPrefix(s[3], "\""), "\"")
		data = append(data, show)
		if line[:3] == "END" {
			break
		}
	}
	return data, nil
}

func (conn *connection) commandExpect(input, expected string) (string, error) {
	result, err := conn.command(input)
	if err != nil {
		return result, err
	}
	if result != expected {
		return "", errors.New("Expected " + expected + " but server returned: " + result)
	}
	return result, nil
}

func (conn *connection) getVar(variable string) (string, error) {
	out, err := conn.command("GET VAR " + conn.UPSName + " " + variable)
	if err != nil {
		_ = level.Error(logger).Log("msg", "problem read VAR ["+variable+"]", "error", err, "ups", conn.UPSName, "data", out)
		return out, err
	}
	_ = level.Debug(logger).Log("msg", "success read VAR ["+variable+"]", "data", out)
	return strings.TrimSuffix(strings.TrimPrefix(strings.Split(out, " ")[3], "\""), "\""), nil
}

func (conn *connection) getList(typeName string) (string, error) {
	out, err := conn.commandList("LIST " + typeName + " " + conn.UPSName)
	if err != nil {
		_ = level.Error(logger).Log("msg", "problem read LIST ["+typeName+"]", "error", err, "ups", conn.UPSName)
		return strings.Join(out, "\n"), err
	}
	_ = level.Debug(logger).Log("msg", "success read LIST ["+typeName+"]", "lines", len(out))
	return strings.Join(out, "\n"), nil
}
