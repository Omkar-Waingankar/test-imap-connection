package main

import (
	"fmt"
	"log"

	"github.com/emersion/go-imap"
	"github.com/emersion/go-imap/client"
)

const (
	serverHost = "REPLACE"
	serverPort = "993"
	username   = "REPLACE"
	password   = "REPLACE"
)

func main() {
	// Connect to server
	log.Println("Connecting to server...")
	c, err := client.DialTLS(fmt.Sprintf("%s:%s", serverHost, serverPort), nil)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Connected")
	defer c.Logout()

	// Login
	if err := c.Login(username, password); err != nil {
		log.Fatal(err)
	}
	log.Println("Logged in")

	// List mailboxes
	mailboxes := make(chan *imap.MailboxInfo, 10)
	done := make(chan error, 1)
	go func() {
		done <- c.List("", "*", mailboxes)
	}()

	log.Println("Mailboxes:")
	for m := range mailboxes {
		fmt.Println("* " + m.Name)
	}

	if err := <-done; err != nil {
		log.Fatal(err)
	}

	// Select INBOX
	_, err = c.Select("INBOX", false)
	if err != nil {
		log.Fatal(err)
	}

	// Search for all messages
	criteria := imap.NewSearchCriteria()
	criteria.WithoutFlags = []string{imap.SeenFlag}
	uids, err := c.Search(criteria)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("UIDs of all messages:", uids)

	// Fetch the latest message by its UID
	if len(uids) > 0 {
		seqSet := new(imap.SeqSet)
		seqSet.AddNum(uids[len(uids)-1]) // Get the last UID

		messages := make(chan *imap.Message, 1)
		section := imap.FetchEnvelope
		done = make(chan error, 1)
		go func() {
			done <- c.Fetch(seqSet, []imap.FetchItem{section}, messages)
		}()

		log.Println("Latest message:")
		msg := <-messages
		if msg != nil {
			fmt.Println("Subject:", msg.Envelope.Subject)
		}

		if err := <-done; err != nil {
			log.Fatal(err)
		}
	} else {
		log.Println("No messages found")
	}
}
