package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/smtp"
	"os"
	"strings"
	"time"
)

//
func main() {
	sourceRoot := os.Args[1]
	if sourceRoot == "" {
		fmt.Fprintf(os.Stderr, "expected path to build sources as first argument")
		os.Exit(1)
	}

	var indata struct {
		Source struct {
			SMTP struct {
				Host     string
				Port     string
				Username string
				Password string
			}
			From string
			To   []string
		}
		Params struct {
			Subject       string
			Body          string
			SendEmptyBody bool `json:"send_empty_body"`
		}
	}

	inbytes, err := ioutil.ReadAll(os.Stdin)
	if err != nil {
		panic(err)
	}

	err = json.Unmarshal(inbytes, &indata)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error parsing input as JSON: %s", err)
		os.Exit(1)
	}

	if indata.Source.SMTP.Host == "" {
		fmt.Fprintf(os.Stderr, `missing required field "source.smtp.host"`)
		os.Exit(1)
	}

	if indata.Source.SMTP.Port == "" {
		fmt.Fprintf(os.Stderr, `missing required field "source.smtp.port"`)
		os.Exit(1)
	}

	if indata.Source.SMTP.Username == "" {
		fmt.Fprintf(os.Stderr, `missing required field "source.smtp.username"`)
		os.Exit(1)
	}

	if indata.Source.SMTP.Password == "" {
		fmt.Fprintf(os.Stderr, `missing required field "source.smtp.password"`)
		os.Exit(1)
	}

	if indata.Source.From == "" {
		fmt.Fprintf(os.Stderr, `missing required field "source.from"`)
		os.Exit(1)
	}

	if len(indata.Source.To) == 0 {
		fmt.Fprintf(os.Stderr, `missing required field "source.to"`)
		os.Exit(1)
	}

	if indata.Params.Subject == "" {
		fmt.Fprintf(os.Stderr, `missing required field "params.subject"`)
		os.Exit(1)
	}

	subject := indata.Params.Subject

	headers := "MIME-version: 1.0\nContent-Type: text/html; charset='UTF-8'"

	body := ""
	if indata.Params.Body != "" {
		body = indata.Params.Body
	}

	type MetadataItem struct {
		Name  string
		Value string
	}
	var outdata struct {
		Version struct {
			Time time.Time
		} `json:"version"`
		Metadata []MetadataItem
	}
	outdata.Version.Time = time.Now().UTC()
	outdata.Metadata = []MetadataItem{
		{Name: "smtp_host", Value: indata.Source.SMTP.Host},
		{Name: "subject", Value: subject},
	}
	outbytes, err := json.Marshal(outdata)
	if err != nil {
		panic(err)
	}

	var messageData []byte
	messageData = append(messageData, []byte("To: "+strings.Join(indata.Source.To, ", ")+"\n")...)
	messageData = append(messageData, []byte("From: "+indata.Source.From+"\n")...)
	if headers != "" {
		messageData = append(messageData, []byte(headers+"\n")...)
	}
	messageData = append(messageData, []byte("Subject: "+subject+"\n")...)

	messageData = append(messageData, []byte("\n")...)
	messageData = append(messageData, []byte(body)...)

	if indata.Params.SendEmptyBody == false && len(body) == 0 {
		fmt.Fprintf(os.Stderr, "Message not sent because the message body is empty and send_empty_body parameter was set to false. Github readme: https://github.com/jointeffort/email-resource")
		fmt.Printf("%s", []byte(outbytes))
		return
	}

	err = smtp.SendMail(
		fmt.Sprintf("%s:%s", indata.Source.SMTP.Host, indata.Source.SMTP.Port),
		smtp.PlainAuth(
			"",
			indata.Source.SMTP.Username,
			indata.Source.SMTP.Password,
			indata.Source.SMTP.Host,
		),
		indata.Source.From,
		indata.Source.To,
		messageData,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Unable to send an email using SMTP server %s on port %s: %v",
			indata.Source.SMTP.Host, indata.Source.SMTP.Port, err)
		os.Exit(1)
	}

	fmt.Printf("%s", []byte(outbytes))
}
