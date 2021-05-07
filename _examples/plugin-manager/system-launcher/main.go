package main

import (
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func panicOnErr(err error) {
	if err != nil {
		log.Panic(err)
	}
}

func main() {
	binpath, err := filepath.Abs(os.Args[0])
	panicOnErr(err)

	log.Println(binpath)

	execPath := filepath.Join(filepath.Dir(os.Args[0]), "PsExec.exe")
	execAbsPath, err := filepath.Abs(execPath)
	panicOnErr(err)

	log.Println(execAbsPath)

	pmPath := filepath.Join(filepath.Dir(os.Args[0]), "..", "plugin-manager.exe")
	pmAbsPath, err := filepath.Abs(pmPath)
	panicOnErr(err)
	log.Println(pmAbsPath)

	log.Println(exec.Command(execAbsPath, "-i", "-s", pmAbsPath).Run())
}
