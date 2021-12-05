package main

import (
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"os/exec"
	"strconv"
	"strings"

	"gopkg.in/yaml.v2"
)

type command struct {
	Name string   `yaml:"name"`
	Type string   `yaml:"type"`
	Cmd  []string `yaml:"cmd"`
}

type runner struct {
	Name        string    `yaml:"name"`
	Description string    `yaml:"description"`
	Commands    []command `yaml:"commands"`
}

func getRunner(filename string) (*runner, error) {
	yamlFile, err := ioutil.ReadFile(filename)
	if err != nil {
		log.Printf("Problem reading file #%v ", err)
		return nil, err
	}

	response := &runner{}
	err = yaml.Unmarshal(yamlFile, response)
	if err != nil {
		log.Fatalf("Problem with Unmarshal: %v", err)
	}

	return response, nil
}

func runCommand(command string, commandWithArgs []string) (response []byte, err error) {
	fmt.Println("Type:", command)
	switch command {
	case "shell":
		return runShellCommand(commandWithArgs)
	case "containerManagement":
		return runContainerManagementCommand(commandWithArgs)
	default:
		return nil, errors.New("Command '" + command + "' is not known")
	}
}

func runShellCommand(commandWithArgs []string) (response []byte, err error) {
	/*
		A possible other way:
		var stderr bytes.Buffer
		cmd := exec.Command("bash", "-c", commandWithArgs...)
		cmd.Stderr = &stderr
		stdout, err := cmd.Output()
		return err, stdout, stderr
	*/
	commandsLength := len(commandWithArgs)
	if commandsLength == 1 {
		command := commandWithArgs[0]
		fmt.Println("Shell Command:", command)
		cmd := exec.Command(command)
		stdout, err := cmd.Output()
		return stdout, err
	} else if commandsLength > 1 {
		command := commandWithArgs[0]
		arguments := commandWithArgs[1:commandsLength]
		fmt.Println("Shell Command:", command)
		fmt.Println("Arguments:", arguments)
		cmd := exec.Command(command, arguments...)

		stdout, err := cmd.Output()
		return stdout, err
	}

	return nil, errors.New("can't process empty shell command")
}

func runContainerManagementCommand(commandWithArgs []string) (response []byte, err error) {
	commandsLength := len(commandWithArgs)
	if commandsLength == 1 {
		command := commandWithArgs[0]
		fmt.Println("Container Command:", command)
		switch command {
		case "ps":
			command := "docker"
			arguments := "ps"
			cmd := exec.Command(command, arguments)
			stdout, err := cmd.Output()
			return stdout, err
		default:
			return nil, fmt.Errorf("command '%s' is not currently supported", command)
		}

	} else if commandsLength > 1 {
		command := "docker"
		arguments := commandWithArgs
		fmt.Println("Container Command:", command)
		fmt.Println("Arguments:", arguments)

		cmdStr := fmt.Sprintf("%s %s", command, strings.Join(arguments, " "))
		fmt.Println("Full Command:", cmdStr)
		cmd := exec.Command(command, arguments...)
		stdout, err := cmd.Output()
		return stdout, err
	}

	return nil, errors.New("can't process empty container command")
}

func main() {
	updateRunner, err := getRunner("sample.yaml")
	if err != nil {
		log.Fatal(err)
	}

	//jsonified, _ := json.Marshal(updateRunner)
	//fmt.Println(string(jsonified))

	fmt.Println("Running " + updateRunner.Name + "\n  " + updateRunner.Description)
	if len(updateRunner.Commands) == 0 {
		fmt.Println("No commands found")
	}

	for i, command := range updateRunner.Commands {
		var header strings.Builder
		header.WriteString(command.Name)
		header.WriteString(" (")
		header.WriteString(strconv.Itoa(i + 1))
		header.WriteString("/")
		header.WriteString(strconv.Itoa(len(updateRunner.Commands)))
		header.WriteString(")")
		header.WriteString("\n--------------------------")
		fmt.Println(header.String())
		output, err := runCommand(command.Type, command.Cmd)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Output:\n%s\n", output)
	}
}
