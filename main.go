package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/credentials/stscreds"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/iam"
	"github.com/go-ini/ini"
	"github.com/mitchellh/go-homedir"
)

var (
	TaskList   bool
	TaskRotate string
	TaskDelete string
)

func ok(err error) {
	if err != nil {
		log.Fatal(err.Error())
	}
}

func main() {
	flag.BoolVar(&TaskList, "list", false, "List access keys")
	flag.StringVar(&TaskRotate, "rotate", "", "Rotate an access key")
	flag.StringVar(&TaskDelete, "delete", "", "Delete an access key")
	flag.Parse()

	if !TaskList && TaskRotate == "" && TaskDelete == "" {
		log.Fatal("must specify one of -list, -rotate, or -delete")
	}

	if TaskList && (TaskRotate != "" || TaskDelete != "") ||
		TaskRotate != "" && (TaskList || TaskDelete != "") ||
		TaskDelete != "" && (TaskList || TaskRotate != "") {
		log.Fatal("must only specify one of -list, -rotate, -delete")
	}

	cfg := aws.NewConfig().
		WithMaxRetries(5)

	sess := session.Must(session.NewSessionWithOptions(session.Options{
		SharedConfigState:       session.SharedConfigEnable,
		AssumeRoleTokenProvider: stscreds.StdinTokenProvider,
		Config:                  *cfg,
	}))

	svc := iam.New(sess)

	getUserOutput, err := svc.GetUser(&iam.GetUserInput{})
	ok(err)

	user := getUserOutput.User.UserName

	if TaskList {
		listAccessKeysOutput, err := svc.ListAccessKeys(&iam.ListAccessKeysInput{UserName: user})
		ok(err)

		for _, accessKeyMetadata := range listAccessKeysOutput.AccessKeyMetadata {
			accessKeyID := accessKeyMetadata.AccessKeyId
			created := accessKeyMetadata.CreateDate
			status := accessKeyMetadata.Status
			fmt.Println("Access Key ID: " + *accessKeyID)
			fmt.Println("Creation Date: " + created.String())
			fmt.Println("Status: " + *status)
			fmt.Println()
		}
		return
	}

	if TaskRotate != "" {
		id := TaskRotate

		_, err = svc.GetAccessKeyLastUsed(&iam.GetAccessKeyLastUsedInput{
			AccessKeyId: &id,
		})
		if err != nil {
			if awsErr, ok := err.(awserr.Error); ok {
				if awsErr.Code() == "ErrCodeNoSuchEntityException" {
					log.Fatal("the provided aws access key id does not exist")
				}
			}
			log.Fatal("failed to check the provided aws access key id: " + err.Error())
		}

		createAccessKeyOutput, err := svc.CreateAccessKey(&iam.CreateAccessKeyInput{})
		ok(err)
		log.Println("created new access key")

		_, err = svc.UpdateAccessKey(&iam.UpdateAccessKeyInput{
			AccessKeyId: &id,
			Status:      aws.String("Inactive"),
		})
		ok(err)
		log.Printf("marked access key %q as inactive", id)

		newKey := createAccessKeyOutput.AccessKey

		fmt.Println()
		fmt.Println("New credentials")
		fmt.Println("---------------")
		fmt.Println("Access Key ID: " + *newKey.AccessKeyId)
		fmt.Println("Secret Access Key: " + *newKey.SecretAccessKey)
		fmt.Println("Creation Date: " + newKey.CreateDate.String())
		fmt.Println("Status: " + *newKey.Status)

		fmt.Println()
		err = updateCredentialsFile(id, *newKey.AccessKeyId, *newKey.SecretAccessKey)
		if err != nil {
			fmt.Printf("Unable to automatically update credentials file. Error: %s\n", err.Error())
		}
		fmt.Println("Automatically updated credentials file.")

		fmt.Println()
		fmt.Println("After confirming the new credentials work, use -delete [ID] to delete the previous access key")
		return
	}

	if TaskDelete != "" {
		id := TaskDelete
		_, err := svc.DeleteAccessKey(&iam.DeleteAccessKeyInput{
			AccessKeyId: &id,
		})
		ok(err)
		log.Printf("successfully deleted access key %q", id)
		return
	}
}

func updateCredentialsFile(oldAccessKeyID, newAccessKeyID, newSecretAccessKey string) error {
	home, err := homedir.Dir()
	if err != nil {
		return fmt.Errorf("unable to detect user home directory: %w", err)
	}

	credentialsFile := filepath.Clean(filepath.Join(home, ".aws", "credentials"))
	credentialsFileBak := filepath.Clean(filepath.Join(home, ".aws", "credentials.bak"))

	credsFile, err := os.Open(credentialsFile)
	if os.IsNotExist(err) {
		return fmt.Errorf("unable to open credentials file %q: %w", credentialsFile, err)
	}

	fmt.Printf("Found AWS Credentials file: %s\n", credentialsFile)

	bak, err := os.Create(credentialsFileBak)
	if err != nil {
		return fmt.Errorf("unable to create backup file %q: %w", credentialsFileBak, err)
	}

	if _, err := io.Copy(bak, credsFile); err != nil {
		return fmt.Errorf("unable to write backup file %q: %w", credentialsFileBak, err)
	}

	creds, _ := ini.Load(credentialsFile)
	for _, section := range creds.Sections() {
		if section.HasKey("aws_access_key_id") {
			key := section.Key("aws_access_key_id")
			val := key.MustString("")
			if val == oldAccessKeyID {
				key.SetValue(newAccessKeyID)
				secret := section.Key("aws_secret_access_key")
				secret.SetValue(newSecretAccessKey)
				if err := creds.SaveTo(credentialsFile); err != nil {
					return fmt.Errorf("unable to save updated credentials file %q: %w", credentialsFile, err)
				}
				return nil
			}
		}
	}

	return fmt.Errorf("unable to find access_key_id %q in credentials file %q", oldAccessKeyID, credentialsFile)
}
