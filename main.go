package main

import (
	"fmt"
	"log"

	"github.com/eefret/gator/internal/config"
)

func main() {
	// Read the configuration from ~/.gatorconfig.json
	cfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config: %v", err)
	}
	fmt.Printf("Initial config: %+v\n", cfg)

	// Update the current user name in the config and write it back to disk
	if err := cfg.SetUser("myusername"); err != nil {
		log.Fatalf("Error updating user: %v", err)
	}
	fmt.Println("User updated successfully!")

	// Read again to confirm changes
	updatedCfg, err := config.Read()
	if err != nil {
		log.Fatalf("Error reading config after update: %v", err)
	}
	fmt.Printf("Updated config: %+v\n", updatedCfg)
}
