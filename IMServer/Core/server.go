package Core

import (
	"bytes"
	"net"
	"regexp"
	"strings"
	"time"
)

var UserList = make(map[string]map[int]User)	//	在线用户列表
var FailMessageList = make(map[string]map[int]Message)	// 延时信息

func Server(listen string){

	if server,err := net.Listen("tcp",listen);err == nil{

		for{

			if conn,err := server.Accept();err == nil{

				go Runner(conn)

			}

		}
	}else{
		panic(err)
	}
}

func Runner(conn net.Conn) {
	var U User
	defer conn.Close()

	// 注册登录操作
	var buf bytes.Buffer
	var b = make([]byte,1024)
	for{
		// 但出现\r\n时停止读取
		if bytes.Contains(buf.Bytes(),[]byte("\r\n")) {
			break;
		}

		num,err := conn.Read(b)
		buf.Write(b[0:num])

		if err != nil{
			return
		}
	}

	// 检测注册消息
	if arr := regexp.MustCompile("^HELO:([^\r\n \t\b:@$#&]+)\r\n$").FindStringSubmatch(buf.String()); len(arr)>0{

		conn.Write([]byte("AUTH_SUCCEED\r\n"))
		U = User{conn,arr[1]}
		if len(UserList[U.Username]) == 0 {

			// 新用户推送给已在线的用户
			for _,us := range UserList{
				for _,c := range us{
					c.Conn.Write([]byte("REFRESHUSERLIST:\r\n"))
				}
			}

			UserList[U.Username] = make(map[int]User)
		}
		num := len(UserList[U.Username])
		UserList[U.Username][num] = U

		defer func() {
			println(U.Username+"@offline")
			delete(UserList[U.Username],num)	// 当掉线时从在线用户列表中删除该用户
			conn.Close()
		}()

	}else{

		conn.Write([]byte("AUTH_FAILED\r\n"))
		return
	}

	// 处理延时消息
	/*var failMsgLen = len(FailMessageList[U.Username])
	for i := 0; failMsgLen > i;i++ {

		var m = FailMessageList[U.Username][i]
		U.Send(m)
		delete(FailMessageList[U.Username],i)
	}*/

	// 接收客户端消息发送操作

	for{


		// 读取数据
		for{
			// 但出现\r\n时停止读取
			if bytes.Contains(buf.Bytes(),[]byte("\r\n")) {
				break;
			}

			num,err := conn.Read(b)
			buf.Write(b[0:num])

			if err != nil{
				conn.Close()
				return
			}
		}

		if arr := regexp.MustCompile("^([^\n\r :@]+)@([^\r\n]+)\r\n$").FindStringSubmatch(buf.String()); len(arr)>0{

			var ToUsernames = strings.Split(arr[1],",")
			var Data = arr[2]
			M := Message{
				Data: Data,
				From: U.Username,
				To: "",
				Time: time.Now().Format("2006-01-02 15:04:05"),
			}
			//U.Conn.Write([]byte("Sent:"+Data+"\r\n"))
			for _,ToUsername := range ToUsernames {
				M.To = ToUsername
				var ToUsers = UserList[M.To]
				for _,ToUser := range ToUsers{

					ToUser.Send(M)
				}
				if len(ToUsers) > 0{
					continue
				}
				if len(FailMessageList[M.To]) == 0 {
					FailMessageList[M.To] = make(map[int]Message)
				}
				FailMessageList[M.To][len(FailMessageList[M.To])] = M

			}

		}else if arr := regexp.MustCompile("^QUIT:([^\r]*)\r\n$").FindStringSubmatch(buf.String()); len(arr)>0{

			return
		}else if arr := regexp.MustCompile("^GETUSERLIST:([^\r]*)\r\n$").FindStringSubmatch(buf.String()); len(arr)>0{

			var userlist = []string{}
			for k,_ := range UserList{

				userlist = append(userlist,k)
			}
			conn.Write([]byte("USERLIST:"+strings.Join(userlist,",")+"\r\n"))

		}else if arr := regexp.MustCompile("^GETMISSMSG:([^\r]*)\r\n$").FindStringSubmatch(buf.String()); len(arr)>0{
			println("获取MISS")
			var failMsgLen = len(FailMessageList[U.Username])
			for i := 0; failMsgLen > i;i++ {

				var m = FailMessageList[U.Username][i]
				U.Send(m)
				delete(FailMessageList[U.Username],i)
				time.Sleep(100)
			}

		}

		buf.Reset()

	}


}
