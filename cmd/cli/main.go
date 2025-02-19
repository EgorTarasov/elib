package main

import (
	"bufio"
	"elib/internal/client"
	"fmt"
	"os"
	"strconv"
)

const LOGIN_URL = "https://lib.msk.misis.ru/elib/login.php"
const PAGE_URL = "https://lib.msk.misis.ru/elib/libs/view.php?id=%d&page=%d&type=large/fast"
const VIEW_URL = "https://lib.msk.misis.ru/elib/view.php?id=%d"

func main() {

	// studentId := 2107095
	// name := "Георгий"
	var studentId int
	var name string

	client := client.NewElibClient()
	fmt.Println("Welcome to elib parser")

	if err := client.LoadCredentials(); err != nil {
		for {
			if studentId != 0 && name != "" {
				break
			}

			reader := bufio.NewReader(os.Stdin)

			fmt.Print("Enter student id: ")

			studentIdStr, err := reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			studentId, err = strconv.Atoi(studentIdStr[:len(studentIdStr)-1])
			if err != nil {
				fmt.Println(err)
				return
			}

			fmt.Print("Enter name: ")

			name, err = reader.ReadString('\n')
			if err != nil {
				fmt.Println(err)
				return
			}

			err = client.Login(studentId, name)
			if err != nil {
				fmt.Println("Can't login, try again")
				studentId = 0
				name = ""
				continue
			}

			// save credentials to config.json
			client.SaveCredentials(studentId, name)
		}
	}

	var command int

	for {
		fmt.Println("1. Download by id")
		fmt.Println("2. Exit")
		fmt.Print("Enter command: ")
		fmt.Scan(&command)
		switch command {
		case 1:
			fmt.Print("Enter id: ")
			var id int
			fmt.Scan(&id)
			err := client.DownloadDocument(id)
			if err != nil {
				fmt.Println(err)
			}
		case 2:
			return
		}
	}

}
