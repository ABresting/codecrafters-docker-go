//go:build linux
package main

import (
	"fmt"
	"os"
	"os/exec"
	"io/ioutil"
	"io"
	"path"
	"syscall"

)

// Usage: your_docker.sh run <image> <command> <arg1> <arg2> ...
func main() {
	// You can use print statements as follows for debugging, they'll be visible when running tests.
	// fmt.Println("Logs from your program will appear here!")

	// it creates a temp directory in the /tmp directory with a random name
	chrootJail, err := ioutil.TempDir("", "")
	if err != nil {
		fmt.Printf("error creating chroot dir: %v", err)
		os.Exit(1)
	}

	command := os.Args[3]
	args := os.Args[4:len(os.Args)]

	err = copyExecutable(chrootJail, command)

	if err != nil {
		fmt.Printf("Error copying executable: %v", err)
		os.Exit(1)
	}

	err = createDevNull(chrootJail)

	if err != nil {
		fmt.Printf("Error creating '/dev/null': %v", err)
		os.Exit(1)
	}

	if err = syscall.Chroot(chrootJail); err != nil {
		fmt.Printf("Chroot error: %v", err)
		os.Exit(1)
	}

	
	cmd := exec.Command(command, args..., )
	// output, err := cmd.Output()
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{
	    Cloneflags: syscall.CLONE_NEWPID,
	}

	err = cmd.Run()
	if exitError, ok := err.(*exec.ExitError); ok {
       	os.Exit(exitError.ExitCode())
    } else if err != nil {
		fmt.Printf("Err: %v", err)
		os.Exit(1)
	}

	// fmt.Printf("Process ID: %d\n", cmd.Process.Pid)
}


func createDevNull(chrootJail_path string) error{
	if err := os.MkdirAll(path.Join(chrootJail_path,"dev"), 0750); err != nil {
		return err
	}

	return ioutil.WriteFile(path.Join(chrootJail_path, "dev", "null"), []byte{}, 0644)
}

func copyExecutable(chrootDir string, executablePath string) error {
	executablePathInChroot := path.Join(chrootDir,executablePath)

	if err := os.MkdirAll(path.Dir(executablePathInChroot), 0750); err != nil{
		return err
	}

	return copyFile(executablePath, executablePathInChroot)
}

func copyFile(sourceFilePath string, destinationDir string)error {
	sourceFileStat , err := os.Stat(sourceFilePath)
	if err != nil{
		return err
	}

	sourceFile, err := os.Open(sourceFilePath)
	if err != nil{
		return err
	}

	defer sourceFile.Close()

	destinationFile, err := os.OpenFile(destinationDir, os.O_RDWR|os.O_CREATE, sourceFileStat.Mode())
	if err != nil {
		return err
	} 
	defer destinationFile.Close()

	_,err = io.Copy(destinationFile, sourceFile)

	return err
}