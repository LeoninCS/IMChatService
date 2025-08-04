package main

import (
	"flag"
	"fmt"
	"io"
	"net"
	"os"
)

type Client struct {
	ServerIp   string
	ServerPort int
	Name       string
	conn       net.Conn
	flag       int
}

func NewClient(serverIp string, serverPort int) *Client {
	client := &Client{
		ServerIp:   serverIp,
		ServerPort: serverPort,
		flag:       999,
	}
	//链接server
	conn, err := net.Dial("tcp", fmt.Sprintf("%s:%d", serverIp, serverPort))
	if err != nil {
		fmt.Println("net.Dial error")
		return nil
	}
	client.conn = conn
	return client
}

var serverIp string
var serverPort int

func init() {
	flag.StringVar(&serverIp, "ip", "127.0.0.1", "设置服务器IP地址(默认是127.0.0.1)")
	flag.IntVar(&serverPort, "port", 8080, "设置服务器端口(默认是8080)")
}

func debug(error string) {
	fmt.Println(error)
}

func (client *Client) DealResponse() {
	//接受服务器数据 ,永久监听
	io.Copy(os.Stdout, client.conn)
}

func (client *Client) menu() bool {
	var flag int
	fmt.Println("1.公聊模式")
	fmt.Println("2.私聊模式")
	fmt.Println("3.在线人数")
	fmt.Println("4.更新用户名")
	fmt.Println("0.退出")
	fmt.Scanln(&flag)

	if flag >= 0 && flag <= 4 {
		client.flag = flag
		return true
	} else {
		fmt.Println(">>>>>请输入合法范围内的数字<<<<<")
		return false
	}
}

func (client *Client) UpdateName() bool {
	fmt.Println(">>>>>请输入用户名:")
	fmt.Scanln(&client.Name)
	sendMsg := "rename:" + client.Name + "\n"
	_, err := client.conn.Write([]byte(sendMsg))
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return false
	}
	return true
}

func (client *Client) GroupChat() {
	fmt.Println(">>>>>公聊模式(输入exit退出)<<<<<")
	for {
		var msg string
		fmt.Scanln(&msg)
		if msg == "exit" {
			break
		}
		sendMsg := msg + "\n"
		_, err := client.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("conn.Write err:", err)
			break
		}
	}
}

func (client *Client) OneChat() {
	fmt.Println(">>>>>私聊模式(输入exit退出)<<<<<")
	client.Online()
	fmt.Println("请输入对方用户名")
	var toname string
	fmt.Scanln(&toname)

	for {
		var msg string
		fmt.Scanln(&msg)
		if msg == "exit" {
			break
		}
		sendMsg := "to:" + toname + ":" + msg + "\n"
		_, err := client.conn.Write([]byte(sendMsg))
		if err != nil {
			fmt.Println("conn.Write err:", err)
			break
		}
	}
}

func (client *Client) Online() {
	sendMsg := "whoisonline"
	_, err := client.conn.Write([]byte(sendMsg + "\n"))
	if err != nil {
		fmt.Println("conn.Write err:", err)
	}
}

func (client *Client) Run() {
	client.UpdateName()
	for client.flag != 0 {
		for client.menu() != true {
		}

		switch client.flag {
		case 1:
			client.GroupChat()
		case 2:
			client.OneChat()
		case 3:
			client.Online()
		case 4:
			renameFlag := client.UpdateName()
			if renameFlag {
				fmt.Println("更新名字成功")
			} else {
				fmt.Println("更新名字失败")
			}
		}
	}
}

func main() {

	flag.Parse()

	client := NewClient(serverIp, serverPort)
	if client == nil {
		fmt.Println(">>>>>链接服务器失败...")
		return
	}
	go client.DealResponse()

	fmt.Println(">>>>>链接服务器成功...")
	client.Run()

}
