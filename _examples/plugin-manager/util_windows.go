package main

import (
	"log"
	"os/exec"
	"strings"
)

func execute(cwd, cmdpath string, argv ...string) ([]byte, error) {
	log.Printf("%s is excuted at %s %v\n", cmdpath, cwd, argv)
	cmd := exec.Command(cmdpath, argv...)
	cmd.Dir = cwd
	output, err := cmd.CombinedOutput()
	utf8, err := GbkToUtf8(output)
	log.Println(string(utf8), err)

	return output, err
}

func execNetsh(argv []string) bool {
	_, err := execute("", "netsh.exe", argv...)
	if _, ok := err.(*exec.ExitError); ok {
		return false
	}

	return true
}

func allowFirewall(name string, proc string) bool {
	argstr := "advfirewall firewall add rule name=\"" + name + "\" dir=in action=allow"
	argv := strings.Split(argstr, " ")
	argv = append(argv, "program="+proc)
	return execNetsh(argv)
}

func checkFirewall(name string) bool {
	argstr := "advfirewall firewall show rule name=" + name + "verbose"
	argv := strings.Split(argstr, " ")

	return execNetsh(argv)
}

func removeFirewall(name string) bool {
	argstr := "advfirewall firewall delete rule name=\"" + name + "\""
	argv := strings.Split(argstr, " ")

	return execNetsh(argv)
}

func _ensureFireWall(name, binPath string) error {
	removeFirewall(name)
	allowFirewall(name, binPath)
	return nil
}
