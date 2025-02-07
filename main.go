package main

import (
	"context"
	"database/sql"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"github.com/eefret/gator/external/rss"
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
	commands.Register("agg", handleAgg)
	commands.Register("addfeed", middlewareLoggedIn(handleAddFeed))
	commands.Register("feeds", handleFeeds)
	commands.Register("follow", middlewareLoggedIn(handleFollow))
	commands.Register("following", middlewareLoggedIn(handleFollowing))
	commands.Register("unfollow", middlewareLoggedIn(handleUnfollow))
	commands.Register("browse", middlewareLoggedIn(handleBrowse))

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

func middlewareLoggedIn(handler func(s *State, cmd Command, user database.User) error) func(*State, Command) error {
	return func(s *State, cmd Command) error {
		if s.Config.CurrentUserName == "" {
			return fmt.Errorf("You must be logged in to run this command")
		}

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		user, err := s.db.GetUser(ctx, s.Config.CurrentUserName)
		if err != nil {
			return fmt.Errorf("Error getting user: %v", err)
		}

		return handler(s, cmd, user)
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

func handleAgg(s *State, cmd Command) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("Agg just requires one argument")
	}

	timeBetweenRequests, err := time.ParseDuration(cmd.Arguments[0])
	if err != nil {
		return fmt.Errorf("Error parsing duration: %v", err)
	}

	fmt.Println("Collecting feeds every " + timeBetweenRequests.String())

	ticker := time.NewTicker(timeBetweenRequests)

	for ; ; <-ticker.C {
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
		err := scrapeFeeds(ctx, s.db)
		cancel()

		if err != nil {
			fmt.Printf("Error scraping feeds: %v\n", err)
		}

	}

}

func handleAddFeed(s *State, cmd Command, user database.User) error {
	if len(cmd.Arguments) != 2 {
		return fmt.Errorf(`AddFeed requires name and url arguments. example addfeed "<name>" "<url>"`)
	}

	feedName := cmd.Arguments[0]
	feedURL := cmd.Arguments[1]

	// Get current user
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	feed, err := s.db.CreateFeed(ctx, database.CreateFeedParams{
		ID: uuid.New(),
		UserID: user.ID,
		Name: feedName,
		Url: feedURL,
	})
	if err != nil {
		return fmt.Errorf("Error creating feed: %v", err)
	}

	_, err = s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error following feed: %v", err)
	}

	return nil
}

func handleFeeds(s *State, cmd Command) error {
	if len(cmd.Arguments) != 0 {
		return fmt.Errorf("Feeds doesnt allow commands")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	feeds, err := s.db.GetFeeds(ctx)
	if err != nil {
		return fmt.Errorf("Error getting feeds: %v", err)
	}

	for _, feed := range feeds {
		user, err := s.db.GetUserById(ctx, feed.UserID)
		if err != nil {
			return fmt.Errorf("Error getting user: %v", err)
		}

		fmt.Printf("* FeedTitle: %s | FeedURL: (%s) | UserName: %s\n", feed.Name, feed.Url, user.Name)
	}

	return nil
}

func handleFollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("Follow requires one argument")
	}

	feedURL := cmd.Arguments[0]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	feed, err := s.db.GetFeedByURL(ctx, feedURL)
	if err != nil {
		return fmt.Errorf("Error getting feed: %v", err)
	}

	row, err := s.db.CreateFeedFollow(ctx, database.CreateFeedFollowParams{
		UserID: user.ID,
		FeedID: feed.ID,
	})
	if err != nil {
		return fmt.Errorf("Error following feed: %v", err)
	}

	fmt.Printf("%s is now following %s\n", row.UserName, row.FeedName)

	return nil
}

func handleFollowing(s *State, cmd Command, user database.User) error {
	if len(cmd.Arguments) != 0 {
		return fmt.Errorf("Following doesnt allow commands")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	follows, err := s.db.GetFeedFollowsForUser(ctx, user.ID)
	if err != nil {
		return fmt.Errorf("Error getting follows: %v", err)
	}

	for _, follow := range follows {
		fmt.Printf("* %s is following %s\n", follow.UserName, follow.FeedName)
	}

	return nil
}

func handleUnfollow(s *State, cmd Command, user database.User) error {
	if len(cmd.Arguments) != 1 {
		return fmt.Errorf("Unfollow requires one argument")
	}

	feedURL := cmd.Arguments[0]

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := s.db.DeleteFeedFollow(ctx, database.DeleteFeedFollowParams{
		UserID: user.ID,
		Url: feedURL,
	})
	if err != nil {
		return fmt.Errorf("Error unfollowing feed: %v", err)
	}

	fmt.Printf("%s is no longer following %s\n", user.Name, feedURL)

	return nil
}

func handleBrowse(s *State, cmd Command, user database.User) error {
	limit := 2
	if len(cmd.Arguments) == 1 {
		if specifiedLimit, err := strconv.Atoi(cmd.Arguments[0]); err == nil {
			limit = specifiedLimit
		} else {
			return fmt.Errorf("invalid limit: %w", err)
		}
	}

	posts, err := s.db.GetPostsForUser(context.Background(), database.GetPostsForUserParams{
		UserID: user.ID,
		Limit:  int32(limit),
	})
	if err != nil {
		return fmt.Errorf("couldn't get posts for user: %w", err)
	}

	fmt.Printf("Found %d posts for user %s:\n", len(posts), user.Name)
	for _, post := range posts {
		fmt.Printf("%s from %s\n", post.PublishedAt.Time.Format("Mon Jan 2"), post.FeedName)
		fmt.Printf("--- %s ---\n", post.Title)
		fmt.Printf("    %v\n", post.Description)
		fmt.Printf("Link: %s\n", post.Url)
		fmt.Println("=====================================")
	}

	return nil
}

func scrapeFeeds(ctx context.Context, db *database.Queries) error {
	next, err := db.GetNextFeedToFetch(ctx)
	if err != nil {
		return fmt.Errorf("Error getting next feed to fetch: %v", err)
	}

	fmt.Println("Found a feed to fetch!")

	return scrapeFeed(ctx, db, next)
}


func scrapeFeed(ctx context.Context, db *database.Queries, feed database.Feed) error {
	err := db.MarkFeedFetched(ctx, feed.ID)
	if err != nil {
		return fmt.Errorf("Error marking feed fetched: %v", err)
	}

	feedData, err := rss.FetchFeed(ctx, feed.Url)
	if err != nil {
		return fmt.Errorf("Error fetching feed: %v", err)
	}

	for _, item := range feedData.Channel.Item {
		publishedAt := sql.NullTime{}
		if t, err := time.Parse(time.RFC1123Z, item.PubDate); err == nil {
			publishedAt = sql.NullTime{
				Time:  t,
				Valid: true,
			}
		}

		_, err = db.CreatePost(context.Background(), database.CreatePostParams{
			FeedID:    feed.ID,
			Title:     item.Title,
			Description: sql.NullString{
				String: item.Description,
				Valid:  true,
			},
			Url:         item.Link,
			PublishedAt: publishedAt,
		})

		if err != nil {
			return fmt.Errorf("Error creating post: %v", err)
		}
	}

	return nil
}