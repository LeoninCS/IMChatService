package main

import (
	"fmt"
	"io"
	"net"
	"sync"
	"time"
)

type Server struct {
	Ip   string
	Port int

	//用户列表
	OnlineMap map[string]*User
	mapLock   sync.RWMutex

	//消息广播的channel
	Message chan string
}

// 监听广播消息的channel
func (this *Server) ListenMessager() {
	for {
		msg := <-this.Message

		this.mapLock.Lock()
		for _, cli := range this.OnlineMap {
			cli.C <- msg
		}
		this.mapLock.Unlock()
	}
}

// 广播业务
func (this *Server) BroadCast(user *User, msg string) {
	sendMsg := "[" + user.Addr + "]" + user.Name + ":" + msg
	this.Message <- sendMsg
}

func (this *Server) Handler(conn net.Conn) {
	//当前业务
	fmt.Println(">>>>>链接建立成功...")

	user := NewUser(conn, this)
	user.Online()
	//监听是否活跃
	isLive := make(chan bool)
	go func() {
		buf := make([]byte, 1024)
		for {
			n, err := conn.Read(buf)

			if n == 0 {
				user.Offline()
				return
			}
			if err != nil && err != io.EOF {
				fmt.Println("Conn Read err:", err)
				return
			}
			msg := string(buf[:n-1])
			user.DoMessage(msg)
			isLive <- true
		}
	}()

	inactivityTimeout := 10 * time.Minute
	timer := time.NewTimer(inactivityTimeout)

	for {
		select {
		case <-isLive:
			if !timer.Stop() {
				<-timer.C
			}
			timer.Reset(inactivityTimeout)
		case <-timer.C:
			user.SendMsg("您长时间未发消息，已被移出服务器")
			this.mapLock.Lock()
			delete(this.OnlineMap, user.Name)
			this.mapLock.Unlock()
			close(user.C)
			user.conn.Close()
			return
		}
	}

}

func NewServer(ip string, port int) *Server {
	server := &Server{
		Ip:        ip,
		Port:      port,
		OnlineMap: make(map[string]*User),
		Message:   make(chan string),
	}
	return server
}

func (this *Server) Start() {
	//listen
	listener, err := net.Listen("tcp", fmt.Sprintf("%s:%d", this.Ip, this.Port))
	if err != nil {
		fmt.Println("net.Listen err:", err)
		return
	}
	//close listen socket
	defer listener.Close()

	//启动监听
	go this.ListenMessager()
	//accept
	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("listener accept err:", err)
			continue
		}
		//do handler
		go this.Handler(conn)
	}

}
