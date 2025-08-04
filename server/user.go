package main

import (
	"net"
)

type User struct {
	Name   string
	Addr   string
	C      chan string
	conn   net.Conn
	server *Server
}

func NewUser(conn net.Conn, server *Server) *User {
	userAddr := conn.RemoteAddr().String()

	user := &User{
		Name:   userAddr,
		Addr:   userAddr,
		C:      make(chan string),
		conn:   conn,
		server: server,
	}

	go user.ListenMessage()

	return user
}

// 上线
func (this *User) Online() {

	//用户上线
	this.server.mapLock.Lock()
	this.server.OnlineMap[this.Name] = this
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "已上线")
}

// 下线
func (this *User) Offline() {

	//用户上线
	this.server.mapLock.Lock()
	delete(this.server.OnlineMap, this.Name)
	this.server.mapLock.Unlock()

	this.server.BroadCast(this, "下线")
}

// 向客户端发送消息
func (this *User) SendMsg(msg string) {
	this.conn.Write([]byte(msg + "\n"))
}

// 广播
func (this *User) DoMessage(msg string) {
	idx := len(msg)
	for i := 0; i < len(msg); i++ {
		if msg[i] == ':' {
			idx = i
			break
		}
	}
	sign := msg[:idx]
	content := msg[min(idx+1, len(msg)):]
	if sign == "whoisonline" {
		this.server.mapLock.Lock()
		for _, user := range this.server.OnlineMap {
			onlineMsg := "[" + user.Addr + "] " + user.Name + ":在线..."
			this.SendMsg(onlineMsg)
		}
		this.server.mapLock.Unlock()
	} else if sign == "rename" {
		_, ok := this.server.OnlineMap[content]
		if ok {
			renameMsg := "该名字已被使用"
			this.SendMsg(renameMsg)
		} else {
			this.server.mapLock.Lock()
			delete(this.server.OnlineMap, this.Name)
			this.Name = content
			this.server.OnlineMap[this.Name] = this
			this.server.mapLock.Unlock()
			renameMsg := "已成功修改用户名为:" + this.Name
			this.SendMsg(renameMsg)
		}

		this.Name = content
	} else if sign == "to" {
		idxx := -1
		for i := 0; i < len(content); i++ {
			if content[i] == ':' {
				idxx = i
				break
			}
		}
		if idxx == -1 {
			toMsg := "发送失败，格式为to:name:message"
			this.SendMsg(toMsg)
		} else {
			name := content[:idxx]
			message := content[min(idxx+1, len(content)):]
			toUser, ok := this.server.OnlineMap[name]
			if !ok {
				toMsg := "该用户不存在"
				this.SendMsg(toMsg)
			} else {
				toUser.SendMsg(this.Name + ":" + message)
			}
		}
	} else {
		this.server.BroadCast(this, msg)
	}

}

// 收到消息转发到客户端
func (this *User) ListenMessage() {
	for {
		msg := <-this.C
		this.SendMsg(msg)
	}
}
