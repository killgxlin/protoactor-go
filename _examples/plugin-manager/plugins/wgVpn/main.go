package main

import (
	"bytes"
	"fmt"
	// "io/ioutil"
	// "net/http"
	// "net/url"
	// "log"
	"os"
	"strconv"
	"net"
	"io"
	"runtime"
	"path/filepath"
	"os/exec"
	"bufio"
	"strings"

	// "github.com/robfig/cron"
)

var port int = 0
var connToServer net.Conn 

func getVpnHomePath() string {
	dir, _ := filepath.Abs(filepath.Dir(os.Args[0]))
	return dir
}

func runSystem(cmdType, cmdpath string, argv ...string) error {
	cmd := exec.Command(cmdpath, argv...)
	var ret error

	go func() {
		stdoutIn, _ := cmd.StdoutPipe()
		defer stdoutIn.Close()
		var stdoutBuf = bufio.NewReader(stdoutIn)

		err := cmd.Start()
		if err != nil {
			ret = err
			sendMsgToServer(fmt.Sprintf("error:%s argv %v err %v\n", cmdpath, argv, err))
		}
		sendMsgToServer(fmt.Sprintf("logger:%s argv %v run ok!!\n", cmdpath, argv))
	
		var buffer []byte = make([]byte, 4096)
		for{
			n, err := stdoutBuf.Read(buffer)
			if err != nil{
				if err == io.EOF{
					if cmdType == "status" {
						sendMsgToServer(fmt.Sprintf("data:ip:%s|way:%s|key:%s|loginTime:%s\n", ip, way, key, loginTime))
					}
					break
				}else{

				}
			}else{
				processRecvInfo(cmdType, string(buffer[:n]))
			}
		}
	}()
	return ret
}

func zhToUnicode(raw []byte) ([]byte, error) {
	str, err := strconv.Unquote(strings.Replace(strconv.Quote(string(raw)), `\\u`, `\u`, -1))
	if err != nil {
		return nil, err
	}
	return []byte(str), nil
}
func tranString(src string) string {
	des := strconv.QuoteToASCII(src)
	ret := ""

	info := strings.Split(des, "\\")
	for i:=0; i<len(info); i++ {
		if info[i][0] == 'u' {
			ret += "\\" + info[i]
		}
	}

	b, _ := zhToUnicode([]byte(ret))
	return string(b)
}

var statusVpn string 
var ip string 
var way string 
var key string 
var loginTime string 
func processRecvInfo(cmdType, buf string) {
	if cmdType == "status" {
		info := strings.Split(buf, "\n")
		for i:=0; i<len(info); i++ {
			index := strings.Index(info[i], "状态:")
			if index != -1 {
				end := strings.Index(info[i], "登录方式:")
				if end == -1 {
					statusVpn = strings.TrimSpace(info[i][index+len("状态:")+1:])
				} else {
					statusVpn = strings.TrimSpace(info[i][index+len("状态:")+1:end-1])
					endKey := strings.Index(info[i], "公钥:")
					if endKey != -1 {
						way = strings.TrimSpace(info[i][end+len("登录方式:")+1:endKey-1])
					}
				}
				statusVpn = tranString(statusVpn)
				sendMsgToServer(fmt.Sprintf("state:%s\n", statusVpn))
				// status := "IDLE"
				// if statusVpn == "在线" {
				// 	status = "OnLine"
				// } else if statusVpn == "空闲" {
				// 	status = "IDLE"
				// } else if statusVpn == "下线" {
				// 	status = "OffLine"
				// }
				// sendMsgToServer(fmt.Sprintf("state:%s\n", status))
			} else if strings.Index(info[i], "PublicKey:") != -1 {
				x := strings.Split(info[i], ":")
				key = strings.TrimSpace(x[1])
			} else if strings.Index(info[i], "Endpoint:") != -1 {
				x := strings.Split(info[i], ":")
				ip = strings.TrimSpace(x[1])
			} else if strings.Index(info[i], "在线时间:") != -1 {
				x := strings.Split(info[i], ":")
				loginTime = strings.TrimSpace(info[i][len(x[0])+1:])
			}
		}
	} else if cmdType == "close" {
		if buf == "已退出" {
			os.Exit(0)
		}
	} else {
		sendMsgToServer(fmt.Sprintf("logger:%s\n", buf))
	}
}

// var crontab *cron.Cron = nil
// func startQuery(time int) {
// 	crontab = cron.New(cron.WithSeconds()) //精确到秒

// 	spec := fmt.Sprintf("*/%d * * * * ?", time) //cron表达式，每秒一次
// 	crontab.AddFunc(spec, func() {
// 		process("status")
// 	})

// 	crontab.Start()
// }

func installDriver() bool {
	if runtime.GOOS == "windows"{
		drvinst := getVpnHomePath() + "\\wintun-driver\\drvinst.exe"
		err := runSystem("", drvinst, "-i", "wintun.inf", "Wintun")
		if err == nil {
			return true
		}
	}
	return false
}

func startMiVPN() bool {
	if runtime.GOOS == "windows"{
		clientd := getVpnHomePath() + "\\wg-clientd-win.exe"
		err := runSystem("", clientd, "--log-level", "debug", "wg0")
		if err == nil {
			return true
		}
	} else if runtime.GOOS == "linux" {
		clientd := getVpnHomePath() + "\\wg-clientd-linux"
		err := runSystem("", clientd, "--log-level", "debug", "wg0")
		if err == nil {
			return true
		}
	} else if runtime.GOOS == "darwin" {
		clientd := getVpnHomePath() + "\\wg-clientd-mac"
		err := runSystem("", clientd, "--log-level", "debug", "utun7")
		if err == nil {
			return true
		}
	}
	return false
}

func getCli() string {
	if runtime.GOOS == "windows"{
		return getVpnHomePath() + "\\wg-client-cli-win.exe"
	} else if runtime.GOOS == "linux" {
		return getVpnHomePath() + "\\wg-client-cli-linuxc"
	} else if runtime.GOOS == "darwin" {
		return getVpnHomePath() + "\\wg-client-cli-mac"
	} 
	
	return ""
}

func process(cmd string) bool {
	sendMsgToServer("logger:Recv:"+cmd)
	if cmd == "init"{
		if !installDriver(){
			sendMsgToServer("error:installDriver error")
			return false
		}
		if !startMiVPN(){
			sendMsgToServer("error:startMiVPN error")
			return false
		}
		sendMsgToServer("state:inited")
		// startQuery(15)
		return true
	} else {
		client := getCli()
		if client == "" {
			sendMsgToServer("error:Does not support operating system")
			return false
		}

		// s := strings.Split(cmd, " ")
		s := strings.SplitN(cmd, "##", 2)
		if s[0] == "start" {
			err := runSystem("", client, "up", "tgc", "--value", s[1])
			if err == nil {
				return true
			}
		} else if s[0] == "stop" {
			err := runSystem("", client, "down")
			if err == nil {
				return true
			}
		} else if s[0] == "status" {
			err := runSystem("status", client, "status")
			if err == nil {
				return true
			}
		} else if s[0] == "?close?" {
			err := runSystem("status", client, "exit")
			if err == nil {
				return true
			}
		}
		return false
	}
}

func readFully(conn net.Conn) ([]byte, error) {
    // 读取所有响应数据后主动关闭连接
    defer conn.Close()
    result := bytes.NewBuffer(nil)
    var buf [512]byte
    for {
        n, err := conn.Read(buf[0:])

		process(string(buf[0:n]))
        result.Write(buf[0:n])
        if err != nil {
            if err == io.EOF {
                break
            }
            return nil, err
        }
    }
    return result.Bytes(), nil
}

func sendMsgToServer(info string) {
	connToServer.Write([]byte(info+"\n"))
	// if connToServer != nil {
	// 	connToServer.Write([]byte(info))
	// }
}

func main(){
	var err error
	connToServer, err = net.Dial("tcp", "127.0.0.1:"+os.Args[2])
	sendMsgToServer(os.Args[1])

	process("init")
	_, err = readFully(connToServer)
    if err != nil {
		sendMsgToServer(fmt.Sprintf("connect error: %v", err))
        os.Exit(1)
    }

    os.Exit(0)
}