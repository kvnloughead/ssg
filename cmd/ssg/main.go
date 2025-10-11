package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/kvnloughead/ssg/internal/builder"
)

func main() {
	// Define subcommands
	buildCmd := flag.NewFlagSet("build", flag.ExitOnError)
	serveCmd := flag.NewFlagSet("serve", flag.ExitOnError)
	newCmd := flag.NewFlagSet("new", flag.ExitOnError)

	// Build command flags
	buildOutput := buildCmd.String(
		"output", "public", "output directory for generated site")
	buildConfig := buildCmd.String(
		"config", "config.yaml", "path to config file")

	// Serve command flags
	servePort := serveCmd.String("port", "8080", "port to serve on")

	// New command flags
	newTitle := newCmd.String("title", "", "post title")

	// Parse command
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	switch os.Args[1] {
	case "build":
		if err := buildCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing command arguments: %v\n", err)
			os.Exit(1)
		}
		if err := builder.Build(*buildConfig, *buildOutput); err != nil {
			fmt.Fprintf(os.Stderr, "Error building site: %v\n", err)
			os.Exit(1)
		}
		fmt.Println("Site built successfully!")

	case "serve":
		if err := serveCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing command arguments: %v\n", err)
			os.Exit(1)
		}
		if err := builder.Serve(*servePort); err != nil {
			fmt.Fprintf(os.Stderr, "Error serving site: %v\n", err)
			os.Exit(1)
		}

	case "new":
		if err := newCmd.Parse(os.Args[2:]); err != nil {
			fmt.Fprintf(os.Stderr, "Error parsing command arguments: %v\n", err)
			os.Exit(1)
		}
		if *newTitle == "" {
			fmt.Fprintln(os.Stderr, "Error: post title is required")
			newCmd.Usage()
			os.Exit(1)
		}
		if err := builder.NewPost(*newTitle); err != nil {
			fmt.Fprintf(os.Stderr, "Error creating post: %v\n", err)
			os.Exit(1)
		}

	default:
		printUsage()
		os.Exit(1)
	}
}

// printUsage prints the usage information
func printUsage() {
	fmt.Println("SSG - Static Site Generator")
	fmt.Println("\nUsage:")
	fmt.Println("  ssg <command> [flags]")
	fmt.Println("\nCommands:")
	fmt.Println("  build    Build the static site")
	fmt.Println("  serve    Serve the site locally")
	fmt.Println("  new      Create a new post")
	fmt.Println("\nFlags:")
	fmt.Println("  build --output <dir>   Output directory (default: public)")
	fmt.Println("  build --config <file>  Config file (default: config.yaml)")
	fmt.Println("  serve --port <port>    Port to serve on (default: 8080)")
	fmt.Println("  new --title <title>    Post title (required)")
}
