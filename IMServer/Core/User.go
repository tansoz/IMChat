package Core

import "net"

type User struct {
	Conn net.Conn	// socket连接
	Username string	// 用户名
}


func (this *User)Send(message Message){

	this.Conn.Write([]byte(message.Time+"#"+message.From+"@"+message.Data+"\r\n"))
}