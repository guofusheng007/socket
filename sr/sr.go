package main

import (
	"bufio"
	"fmt"
	_ "github.com/codyguo/godaemon"
	"log"
	"net"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"
)

// "."表示当前主程序所在的目录，".\\yaml"表示当前目录下的子目录yaml。注意”\\“是转义方式表达，是一个”\“
const logPath = "."

// 全局常量定义
const (
	flagLogin_user         = "c_GuoFS_USER"       //返回给客户端的标识，表示处于用户名称录入阶段。
	flagLogin_pwd          = "c_GuoFS_PASSWORD"   //返回给客户端的标识，表示处于用户密码录入阶段。
	flagLogin_prompt       = "c_GuoFS_GUOFS"      //返回给客户端的标识，表示处于用户已正常进入系统，可以正常使用。
	flagLogin_accpet_True  = "c_GuoFS_USER_TRUE"  //返回给客户端的标识，c_GuoFS_USER_TRUE表示用户和密码正确，
	flagLogin_accpet_False = "c_GuoFS_USER_FALSE" //返回给客户端的标识，c_GuoFS_USER_FALSE表示不正确。
)

// 全局变量定义
var limitGoroutine chan int

// 最大并发数进栈操作
func limit_receice() {
	<-limitGoroutine
}

// 日志写入
func log_Write(filePath string, logContext []string) {
	//设置日志文件路径和名称
	curFile, _ := exec.LookPath(os.Args[0])
	logFile := filePath + string(os.PathSeparator) + filepath.Base(curFile) + "_" + time.Now().Format(time.DateOnly) + ".log"
	//fmt.Println(logFile)

	//打开日志文件，若不存在，就新建，若存在，就追加。
	objLogFile, err := os.OpenFile(logFile, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		fmt.Println(err)
		return
	}
	defer objLogFile.Close()
	//objLogFile.Write([]byte("test"))
	for _, v := range logContext {
		//logTxt := time.Now().Format(time.DateTime) + ";" + v + "\n"
		write := bufio.NewWriter(objLogFile) //写入文件时，使用带缓存的 *Writer
		write.WriteString(v)
		write.Flush()
	}
}

// 保存程序运行过程中的日志
func save_ErrorLog(err error) {
	logTxt := time.Now().Format(time.DateTime) + ";进程意外中止," + err.Error() + "\n"
	log_Write(logPath, []string{logTxt}) //日志写入日志文件
}

// 错误分析
func checkErr(err error) {
	if err != nil {
		log.Fatalln("程序意外中止：", err)
	}
}

// 业务模块
func handleConn(conn net.Conn) {
	//获得客户端的ip信息
	conn.RemoteAddr()
	//ipString := strings.Split(conn.RemoteAddr().String(), ":")
	//发送数据：提示信息
	//topInfo := "Hello," + ipString[0] + ",I am test-GuoFS\n\n"
	//conn.Write([]byte(topInfo))

	//定义变量
	var (
		buf      = make([]byte, 2048) //读取信息时的保存变量
		count    int                  //接收字串的长度
		err      error                //错误返回值
		userInfo = map[string]string{
			"user":   "guofs",
			"passwd": "123456",
		}
	)

	//验证用户帐号
	for { //接受client端无数次的Wirte发送。

		//读取客户端录入的用户名称
		conn.Write([]byte(flagLogin_user))
		count, err = conn.Read(buf)
		if err != nil { //当client端按下“CTRL+C”时，直接退出当前循环体。
			log.Println("子进程意外中止，", err) //日志打印到屏幕
			save_ErrorLog(err)           //日志保存在日志文件中
			<-limitGoroutine             //当客户端中止时，允许通道写入数据，即释放并发数。
			return                       //从当前函数体返回。
		}
		userName := strings.Trim(string(buf[:count]), "\n")
		//fmt.Println("userName:", userName)

		//读取客户端录入的用户密码
		conn.Write([]byte(flagLogin_pwd))
		count, err = conn.Read(buf)
		if err != nil { //当client端按下“CTRL+C”时，直接退出当前循环体。
			log.Println("子进程意外中止，", err) //日志打印到屏幕
			save_ErrorLog(err)           //日志保存在日志文件中
			<-limitGoroutine             //当客户端中止时，允许通道写入数据，即释放并发数。
			return                       //从当前函数体返回。
		}
		passWord := strings.Trim(string(buf[:count]), "\n")
		//fmt.Println("passWord:", passWord)

		//验证用户，若不正确，请重新录入。若正确，进入业务处理部分
		if !((userName == userInfo["user"]) && (passWord == userInfo["passwd"])) {
			conn.Write([]byte(flagLogin_accpet_False))
		} else {
			conn.Write([]byte(flagLogin_accpet_True))
			break
		}
	}

	//业务处理部分
	var userContext string
	for { //接受client端无数次的Wirte发送。
		//取得用户录入信息。只有用户进入flagLogin_prompt提示阶段才可以。
		//conn.Write([]byte(flagLogin_prompt))
		count, err := conn.Read(buf)

		if err != nil { //当client端按下“CTRL+C”时，直接退出当前循环体。
			log.Println("子进程意外中止，", err) //日志打印到屏幕
			save_ErrorLog(err)           //日志保存在日志文件中
			<-limitGoroutine             //当客户端中止时，允许通道写入数据，即释放并发数。
			return                       //从当前函数体返回。
		}

		userContext = strings.Trim(string(buf[:count]), "\n")
		//log.Println("接收的字节数：", len, "，接收的内容：", userContext, ",客户端IP:", conn.RemoteAddr().String())
		//outStr := "您发送的内容是："userContext + ",您的IP:Port :" + conn.RemoteAddr().String() + "\n"
		//conn.Write([]byte(outStr)) //回显。向客户端发送信息

		switch userContext { //在此部分，可以定义一些本系统所需的一些命令。
		case "exit":
			<-limitGoroutine //当客户端正常退出时，允许通道写入数据，即释放并发数。
			return
		case "hello":
			topInfo := `

系统有如下常用命令
-----------------------------------
help		查看帮助
list		命令列表
ver			查看软件版本
exit		退出

`
			conn.Write([]byte(topInfo))
		case "help":
			conn.Write([]byte("help me\n"))
		case "ver":
			conn.Write([]byte("GuoFS v0.1.8\n"))
		default: //业务处理部分。
			//服务端显示
			str_log := "接收的字节数：" + strconv.Itoa(count) + ",接收的数据：" + userContext + ",from：" + conn.RemoteAddr().String()
			fmt.Println(str_log) //日志打印到屏幕
			//日志写入日志文件
			logTxt := time.Now().Format(time.DateTime) + ";" + str_log + "\n"
			log_Write(logPath, []string{logTxt})
			//回显。向客户端发送信息
			str_send := "发送的数据：" + userContext + ",数据源：" + conn.RemoteAddr().String() + "\n"
			conn.Write([]byte(str_send)) //回显。向客户端发送信息
		}

	}
}

func main() {
	//采用通道方式来控制并发数
	var LimitNum = 100                        //定义最大并发数量
	limitGoroutine = make(chan int, LimitNum) //最大并发数为LimitNum+1。采用channel栈方式来控制最大并发数量
	go limit_receice()                        //让通道处于接收状态

	//打开服务侦听
	sr_PortNum := 1234
	sr_protocol := "tcp"
	sr, err := net.Listen(sr_protocol, ":"+strconv.Itoa(sr_PortNum))
	defer sr.Close()
	if err != nil {
		save_ErrorLog(err)          //日志保存在日志文件中
		log.Fatalln("程序意外中止：", err) //日志打印到屏幕，并退出。
	}
	log.Println("服务启动成功,侦听端口：", sr_protocol, "/", sr_PortNum)
	//fmt.Println(reflect.TypeOf(sr))

	//客户端连接。支持并发，即支持从多个client端读取信息。每一个client占用一个Accept。
	for {
		conn, err := sr.Accept() //等待client连接,此时处于等待状态，当有client连接过来时，其后面的代码才会被运行。每一个客户端线程只有一个accpet。
		defer conn.Close()
		if err != nil { //无法完成tcp三次握手而中止
			log.Println("子进程意外中止，", err) //日志打印到屏幕
			save_ErrorLog(err)           //日志保存在日志文件中
			continue                     //跳出当前循环，继续下一个循环。
		}
		limitGoroutine <- 1 //每一个连接成功后向通道limitGoroutine注入一个值，最大注入量为通道的容量。当通道满时就无法注入，达到控制最大并发。
		//fmt.Println(reflect.TypeOf(conn))
		//调用并发进程
		go handleConn(conn)
	}
}

