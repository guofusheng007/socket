package main

import (
	"bufio"
	"fmt"
	"golang.org/x/crypto/ssh/terminal"
	"log"
	"net"
	"os"
	"strings"
	"time"
)

func main() {
	conn, err := net.DialTimeout("tcp", "192.168.3.52:1234", time.Second*10)
	if err != nil {
		log.Fatalln(err)
	}
	defer conn.Close()
	log.Println("连接成功")

	//进入循环体，对来自服务器的返回进行解析。
	var buf = make([]byte, 2048)
	var userInput string

	fmt.Print("Hello,welcome GuoFS。\nver 0.1.8\n\n")

	for {
		//从sr读取信息
		count, err := conn.Read(buf)
		if err != nil {
			break
		}
		//sysReturnContext := strings.Trim(string(buf[:count]), "\n")
		sysReturnContext := string(buf[:count])
		//fmt.Print(sysReturnContext)
		switch sysReturnContext {
		case "c_GuoFS_USER":
			//fmt.Print(sysReturnContext)
			fmt.Print("User : ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			userInput = scanner.Text()
			if len(scanner.Text()) == 0 { //空用户名录入时的处理，即直接回车，也没有空格。
				conn.Write([]byte("c_GuoFS_USER"))
			}
			conn.Write([]byte(userInput))

		case "c_GuoFS_PASSWORD":
			//fmt.Print(sysReturnContext)
			fmt.Print("Password : ")
			tmp, _ := terminal.ReadPassword(int(os.Stdin.Fd()))
			passwd := strings.Trim(string(tmp), "\n")
			if len(passwd) == 0 { //空密码录入时的处理，即直接回车，也没有空格。当为空时，会出错，类似死循环。
				conn.Write([]byte("c_GuoFS_PASSWORD"))
			}
			conn.Write([]byte(passwd))
		case "c_GuoFS_USER_FALSE": //用于网络有延时情况
			//fmt.Print(sysReturnContext)
			fmt.Print("\n密码不正确，请重新录入。\n\n")
		case "c_GuoFS_USER_FALSEc_GuoFS_USER": //用于网络无延时情况
			//fmt.Print(sysReturnContext)
			fmt.Print("\n密码不正确，请重新录入。\n\n")
			fmt.Print("User : ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			userInput = scanner.Text()
			if len(scanner.Text()) == 0 { //空用户名录入时的处理，即直接回车，也没有空格。当为空时，会出错，类似死循环。
				conn.Write([]byte("c_GuoFS_USER"))
			}
			conn.Write([]byte(userInput))
		case "c_GuoFS_USER_TRUE":
			//fmt.Print(sysReturnContext)
			//fmt.Print("密码正确-4--GuoFS >")
			conn.Write([]byte("hello"))
		default:
			fmt.Print(sysReturnContext, "\n")
			fmt.Print("default GuoFS > ")
			scanner := bufio.NewScanner(os.Stdin)
			scanner.Scan()
			userInput = scanner.Text()
			//fmt.Println("len:", len(userInput))
			if len(userInput) != 0 { //空字串处理。当为空时，会出错，类似死循环。
				conn.Write([]byte(userInput))
				if userInput == "exit" {
					fmt.Println("Bye thks.\n")
					return
				}
			} else {
				conn.Write([]byte("hello"))
			}

		}
	}
}

