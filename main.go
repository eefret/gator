package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/eefret/gator/internal/config"
	"github.com/eefret/gator/internal/database"
	"github.com/google/uuid"

	_ "github.com/lib/pq"
)

type State struct {
	Config *config.Config
	db *database.Queries
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

	db, err := sql.Open("postgres", cfg.DbURL)
	if err != nil {
		log.Fatalf("Error opening database: %v", err)
	}

	dbQueries := database.New(db)

	state := &State{
		Config: cfg,
		db: dbQueries,
	}

	// Create a new instance of the commands struct with an initialized map of handler functions.
	commands := &Commands{
		commands: make(map[string]func(*State, Command) error),
	}
	commands.Register("login", handlerLogin)
	commands.Register("register", handlerRegister)
	commands.Register("reset", handleReset)
	commands.Register("users", handleUsers)

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

	name := cmd.Arguments[0]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := s.db.GetUser(ctx, name)
	if err != nil {
		return fmt.Errorf("Error getting user: %v", err)
	}

	if user.Name == "" {
		return fmt.Errorf("User not found")
	}

	s.Config.CurrentUserName = name
	s.Config.SetUser(name)

	fmt.Println("User has been set successfully!")

	return nil
}

func handlerRegister(s *State, cmd Command) error {
	if len(cmd.Arguments) == 0 {
		return fmt.Errorf("No command provided")
	}

	// Get the name of the user from the command arguments.
	name := cmd.Arguments[0]

	// Create a new user in the database using the CreateUser method from the database package.
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	user, err := s.db.CreateUser(ctx, database.CreateUserParams{
		ID:     uuid.New(),
		Name: 	name,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	})
	if err != nil {
		return fmt.Errorf("Error creating user: %v", err)
	}

	// Log the user in by calling the handlerLogin function with the user’s name.
	loginCmd := Command{
		Name: "login",
		Arguments: []string{name},
	}
	if err := handlerLogin(s, loginCmd); err != nil {
		return fmt.Errorf("Error logging in user: %v", err)
	}

	// Print a message that the user was created, and log the user’s data to the console for your own debugging.
	fmt.Printf("User %s has been created successfully!\n", user.Name)
	fmt.Printf("User ID: %s\n", user.ID)
	fmt.Printf("Created At: %s\n", user.CreatedAt)
	fmt.Printf("Updated At: %s\n", user.UpdatedAt)

	return nil
}

func handleReset(s *State, cmd Command) error {
	if len(cmd.Arguments) != 0 {
		return fmt.Errorf("Reset doesnt allow commands")
	}

	s.Config.CurrentUserName = ""
	s.Config.SetUser("")

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.db.ResetUsers(ctx); err != nil {
		return fmt.Errorf("Error resetting users: %v", err)
	}

	return nil
}

func handleUsers(s *State, cmd Command) error {
	if len(cmd.Arguments) != 0 {
		return fmt.Errorf("Users doesnt allow commands")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	users, err := s.db.GetUsers(ctx)
	if err != nil {
		return fmt.Errorf("Error listing users: %v", err)
	}

	for _, user := range users {
		if user.Name == s.Config.CurrentUserName {
			fmt.Printf("* %s (current)\n", user.Name)
		} else {
			fmt.Printf("* %s\n", user.Name)
		}
	}

	return nil
}