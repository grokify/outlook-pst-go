// Command pstinfo displays information about a PST file.
package main

import (
	"flag"
	"fmt"
	"os"

	outlookpst "github.com/grokify/outlook-pst-go"
)

func main() {
	showMessages := flag.Bool("messages", false, "Show messages in each folder")
	maxMessages := flag.Int("max-messages", 10, "Maximum messages to show per folder")
	showAttachments := flag.Bool("attachments", false, "Show attachment info for messages")
	flag.Parse()

	if flag.NArg() < 1 {
		fmt.Fprintf(os.Stderr, "Usage: %s [options] <pst-file>\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "\nOptions:\n")
		flag.PrintDefaults()
		os.Exit(1)
	}

	filename := flag.Arg(0)

	pst, err := outlookpst.Open(filename)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error opening PST: %v\n", err)
		os.Exit(1)
	}
	defer func() {
		if cerr := pst.Close(); cerr != nil {
			fmt.Fprintf(os.Stderr, "Warning: error closing PST: %v\n", cerr)
		}
	}()

	// Print PST info
	fmt.Printf("PST File: %s\n", filename)
	fmt.Printf("Format: %s\n", pst.Format())
	fmt.Printf("Encryption: %s\n", pst.CryptMethod())

	if pst.IsPST() {
		fmt.Println("Type: PST (Personal Storage Table)")
	} else {
		fmt.Println("Type: OST (Offline Storage Table)")
	}

	name, err := pst.Name()
	if err == nil {
		fmt.Printf("Display Name: %s\n", name)
	}

	fmt.Println()

	// Get root folder
	root, err := pst.RootFolder()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting root folder: %v\n", err)
		os.Exit(1)
	}

	// Print folder tree
	fmt.Println("Folder Structure:")
	fmt.Println("-----------------")
	printFolder(root, "", *showMessages, *maxMessages, *showAttachments)
}

func printFolder(folder *outlookpst.Folder, indent string, showMessages bool, maxMessages int, showAttachments bool) {
	name, err := folder.Name()
	if err != nil {
		name = "<error reading name>"
	}

	contentCount, _ := folder.ContentCount()
	unreadCount, _ := folder.UnreadCount()

	fmt.Printf("%s[%s] (%d items, %d unread)\n", indent, name, contentCount, unreadCount)

	if showMessages {
		count := 0
		for msg, err := range folder.Messages() {
			if err != nil {
				fmt.Printf("%s  Error: %v\n", indent, err)
				break
			}

			if count >= maxMessages {
				remaining := int(contentCount) - maxMessages
				if remaining > 0 {
					fmt.Printf("%s  ... and %d more messages\n", indent, remaining)
				}
				break
			}

			subject, _ := msg.Subject()
			sender, _ := msg.SenderName()
			deliveryTime, _ := msg.DeliveryTime()

			fmt.Printf("%s  - %s\n", indent, subject)
			fmt.Printf("%s    From: %s\n", indent, sender)
			if !deliveryTime.IsZero() {
				fmt.Printf("%s    Date: %s\n", indent, deliveryTime.Format("2006-01-02 15:04:05"))
			}

			if showAttachments {
				attachCount, _ := msg.AttachmentCount()
				if attachCount > 0 {
					fmt.Printf("%s    Attachments (%d):\n", indent, attachCount)
					for att, err := range msg.Attachments() {
						if err != nil {
							fmt.Printf("%s      Error: %v\n", indent, err)
							break
						}
						attName, _ := att.Filename()
						attSize, _ := att.Size()
						fmt.Printf("%s      - %s (%d bytes)\n", indent, attName, attSize)
					}
				}
			}

			count++
		}
	}

	// Recurse into subfolders
	for subfolder, err := range folder.Subfolders() {
		if err != nil {
			fmt.Printf("%s  Error listing subfolders: %v\n", indent, err)
			break
		}
		printFolder(subfolder, indent+"  ", showMessages, maxMessages, showAttachments)
	}
}
