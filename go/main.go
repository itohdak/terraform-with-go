package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-exec/tfinstall"
)

func copyFile(srcName, dstName string) error {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()

	dst, err := os.Create(dstName)
	if err != nil {
		return err
	}
	defer dst.Close()

	_, err = io.Copy(dst, src)
	if err != nil {
		return err
	}
	return nil
}
func printContent(srcName string) error {
	src, err := os.Open(srcName)
	if err != nil {
		return err
	}
	defer src.Close()
	buf := make([]byte, 1024)
	for {
		n, err := src.Read(buf)
		if n == 0 {
			break
		}
		if err != nil {
			return err
		}
		fmt.Println(string(buf[:n]))
	}
	return nil
}

func createNewResource(workingDir string) error {
	// create working directory and copy tf template file
	err := os.Mkdir(workingDir, 0775)
	if err != nil {
		return err
	}
	err = copyFile("/home/ec2-user/template/main.tf", filepath.Join(workingDir, "main.tf"))
	if err != nil {
		return err
	}

	tmpDir, err := ioutil.TempDir("", "tfinstall")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	execPath, err := tfinstall.Find(context.Background(), tfinstall.LatestVersion(tmpDir, false))
	if err != nil {
		return err
	}
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return err
	}
	err = tf.Init(context.Background(), tfexec.Upgrade(true))
	if err != nil {
		return err
	}
	err = tf.Apply(context.Background())
	if err != nil {
		return err
	}
	_, err = tf.Show(context.Background())
	if err != nil {
		return err
	}
	return nil
}
func destroyResource(workingDir string) error {
	tmpDir, err := ioutil.TempDir("", "tfinstall")
	if err != nil {
		log.Fatalf("error creating temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	execPath, err := tfinstall.Find(context.Background(), tfinstall.LatestVersion(tmpDir, false))
	if err != nil {
		return err
	}
	tf, err := tfexec.NewTerraform(workingDir, execPath)
	if err != nil {
		return err
	}
	err = tf.Destroy(context.Background())
	if err != nil {
		return err
	}
	return nil
}

func main() {
	// err := printContent("/home/ec2-user/template/main.tf")
	// if err != nil {
	// 	log.Fatalf("error in printContent")
	// }
	userName := "test"
	createNewResource("/home/ec2-user/" + userName)
	// destroyResource("/home/ec2-user/" + userName)
}
