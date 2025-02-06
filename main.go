package main

import (
	"fmt"
	"log"
	"os"

	"github.com/eefret/gator/internal/config"
)

type State struct {
	Config *config.Config
}

type Command struct {
	Name string
	Arguments []string
}

type Commands struct {
	commands map[string]func(*State, Command) error
}

func (c *Commands) Register(name string, f func(*State, Command) error) {
	c.commands[name] = f
}

func (c *Commands) Run(s *State, cmd Command) error {
	if f, ok := c.commands[cmd.Name]; ok {
		return f(s, cmd)
	}

	return fmt.Errorf("Command %s not found", cmd.Name)
}


func main() {
	// Read the configuration from ~/.gatorconfig.json
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}

	state := &State{
		Config: cfg,
	}

	// Create a new instance of the commands struct with an initialized map of handler functions.
	commands := &Commands{
		commands: make(map[string]func(*State, Command) error),
	}
	commands.Register("login", handlerLogin)

	// Use os.Args to get the command-line arguments passed in by the user.
	// The first argument is the name of the program, so we skip it.
	if len(os.Args) < 2 {
		log.Fatalf("No command provided")
	}

	commandName := os.Args[1]
	commandArgs := os.Args[2:]

	c := Command{
		Name: commandName,
		Arguments: commandArgs,
	}

	if err := commands.Run(state, c); err != nil {
		log.Fatalf("Error running command: %v", err)
	}


}

func handlerLogin(s *State, cmd Command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("No command provided")
	}

	s.Config.CurrentUserName = cmd.Arguments[0]
	s.Config.SetUser(s.Config.CurrentUserName)

	fmt.Println("User has been set successfully!")

	return nil
}
