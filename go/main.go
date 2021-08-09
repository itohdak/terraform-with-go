package main

import (
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/hashicorp/terraform-exec/tfexec"
	"github.com/hashicorp/terraform-exec/tfinstall"
	"github.com/labstack/echo"
	"github.com/slack-go/slack"
)

var api *slack.Client

type CreateResourceRequest struct {
	// TODO: these are only examples. please modify to fit your needs.
	Username     string `json:"username"`
	ResourceType string `json:"resource_type"`
}
type DeleteResourceRequest struct {
	// TODO: these are only examples. please modify to fit your needs.
	Username     string `json:"username"`
	ResourceType string `json:"resource_type"`
}

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
	// TODO: post message to slack DM
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
	os.RemoveAll(workingDir)
	return nil
}
func postHandler(c echo.Context) error {
	req := new(CreateResourceRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	// go createNewResource("/home/ec2-user/" + req.Username)
	userID := "UFECAQTHC" // TODO
	if err := sendSlackDM(
		api,
		userID,
		slack.MsgOptionText("This is a DM", true),
	); err != nil {
		return err
	}
	return c.String(http.StatusCreated, fmt.Sprintf("Creating new resource for %s\n", req.Username))
}
func deleteHandler(c echo.Context) error {
	req := new(DeleteResourceRequest)
	if err := c.Bind(req); err != nil {
		return err
	}
	go destroyResource("/home/ec2-user/" + req.Username)
	return c.String(http.StatusOK, fmt.Sprintf("Removing resource for %s\n", req.Username))
}
func sendSlackDM(client *slack.Client, userID string, options slack.MsgOption) error {
	param := new(slack.OpenConversationParameters)
	param.Users = append(param.Users, userID)
	channel, _, _, err := client.OpenConversation(param)
	if err != nil {
		return err
	}
	channelID := channel.GroupConversation.Conversation.ID
	_, _, err = client.PostMessage(channelID, options)
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

	token, ok := os.LookupEnv("SLACK_TOKEN")
	if !ok {
		fmt.Println("Missing SLACK_TOKEN in environment")
		os.Exit(1)
	}
	api = slack.New(
		token,
		slack.OptionDebug(true),
		slack.OptionLog(log.New(os.Stdout, "slack-bot: ", log.Lshortfile|log.LstdFlags)),
	)

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, World!!")
	})
	e.POST("/instance/create", postHandler)
	e.DELETE("/instance/delete", deleteHandler)
	e.Logger.Fatal(e.Start(":8080"))
}
