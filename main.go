package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
)

type myWriter struct {
	extra string
}

func (m myWriter) Write(p []byte) (n int, err error) {
	return os.Stdout.Write(append(p, m.extra...))
}

type user struct {
	Id    string `json:"id"`
	Email string `json:"email"`
	Age   int    `json:"age"`
}

type Arguments map[string]string

func parseArgs() Arguments {
	idFlag := flag.String("id", "", "user id")
	operationFlag := flag.String(
		"operation",
		"",
		"should be \"add\", \"list\", \"findById\" or \"remove\"")
	itemFlag := flag.String("item", "", "user data in json format")
	fileNameFlag := flag.String("fileName", "", "file to store users")

	flag.Parse()

	return Arguments{
		"id":        *idFlag,
		"operation": *operationFlag,
		"item":      *itemFlag,
		"fileName":  *fileNameFlag,
	}
}

func updateUsersFile(fileName string, users []user) error {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_TRUNC, 0644)
	if err != nil {
		return err
	}
	res, err := json.Marshal(users)
	if err != nil {
		return err
	}
	if _, err = file.Write(res); err != nil {
		return err
	}
	if err = file.Close(); err != nil {
		return err
	}
	return nil
}

func Perform(args Arguments, writer io.Writer) error {
	for _, v := range []string{"operation", "fileName"} {
		if args[v] == "" {
			return fmt.Errorf("-%s flag has to be specified", v)
		}
	}

	// Create file if it doesn't exist
	file, err := os.OpenFile(args["fileName"], os.O_RDONLY|os.O_CREATE, 0644)
	if err != nil {
		return err
	}
	err = file.Close()
	if err != nil {
		return err
	}

	content, err := ioutil.ReadFile(args["fileName"])
	if err != nil {
		return err
	}

	var contentObj []user
	if len(content) != 0 {
		if err := json.Unmarshal(content, &contentObj); err != nil {
			return err
		}
	}

	switch args["operation"] {
	case "list":
		if _, err := writer.Write(content); err != nil {
			return err
		}
	case "add":
		if args["item"] == "" {
			return fmt.Errorf("-item flag has to be specified")
		}
		var itemObj user
		if err := json.Unmarshal([]byte(args["item"]), &itemObj); err != nil {
			return err
		}
		for _, v := range contentObj {
			if v.Id == itemObj.Id {
				_, err := writer.Write([]byte(fmt.Sprintf("Item with id %s already exists", itemObj.Id)))
				if err != nil {
					return err
				}
				return nil
			}
		}
		if err := updateUsersFile(args["fileName"], append(contentObj, itemObj)); err != nil {
			return err
		}
	case "findById":
		if args["id"] == "" {
			return fmt.Errorf("-id flag has to be specified")
		}
		for _, v := range contentObj {
			if v.Id == args["id"] {
				res, err := json.Marshal(v)
				if err != nil {
					return err
				}
				if _, err = writer.Write(res); err != nil {
					return err
				}
				return nil
			}
		}
	case "remove":
		if args["id"] == "" {
			return fmt.Errorf("-id flag has to be specified")
		}
		for k, v := range contentObj {
			if v.Id == args["id"] {
				err := updateUsersFile(args["fileName"], append(contentObj[:k], contentObj[k+1:]...))
				if err != nil {
					return err
				}
				return nil
			}
		}
		_, err = writer.Write([]byte(fmt.Sprintf("Item with id %s not found", args["id"])))
		if err != nil {
			return err
		}
	default:
		return fmt.Errorf("Operation %s not allowed!", args["operation"])
	}

	return nil
}

func main() {
	err := Perform(parseArgs(), myWriter{"\n"})
	if err != nil {
		panic(err)
	}
}
