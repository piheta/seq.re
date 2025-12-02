package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/piheta/seq.re/cmd/cli/client"
	"github.com/piheta/seq.re/cmd/cli/commands"
	"github.com/piheta/seq.re/cmd/cli/config"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	command := os.Args[1]

	// Handle version command
	if command == "version" {
		commands.Version(version, commit, date)
		return
	}

	// Handle config commands (don't need server connection)
	if command == "config" {
		if err := handleConfigCommand(); err != nil {
			slog.Error(err.Error())
			os.Exit(1)
		}
		return
	}

	// All other commands need API client
	serverURL := config.GetServerURL()
	apiClient := client.New(serverURL)

	var err error
	switch command {
	case "url":
		if len(os.Args) < 3 {
			_, _ = fmt.Fprint(os.Stdout, "Usage: seqre url <URL>\n")
			os.Exit(1)
		}
		err = commands.URLShorten(apiClient, os.Args[2])

	case "expand":
		if len(os.Args) < 3 {
			_, _ = fmt.Fprint(os.Stdout, "Usage: seqre expand <short>\n")
			os.Exit(1)
		}
		err = commands.URLExpand(apiClient, os.Args[2])

	case "ip":
		err = commands.IP(apiClient)

	case "secret":
		if len(os.Args) < 3 {
			_, _ = fmt.Fprint(os.Stdout, "Usage: seqre secret <text>\n")
			_, _ = fmt.Fprint(os.Stdout, "       seqre secret get <short> <key>\n")
			os.Exit(1)
		}
		if os.Args[2] == "get" {
			if len(os.Args) < 5 {
				_, _ = fmt.Fprint(os.Stdout, "Usage: seqre secret get <short> <key>\n")
				os.Exit(1)
			}
			err = commands.SecretGet(apiClient, os.Args[3], os.Args[4])
		} else {
			// Join all args from index 2 onwards to support multi-word secrets
			secretText := ""
			for i := 2; i < len(os.Args); i++ {
				if i > 2 {
					secretText += " "
				}
				secretText += os.Args[i]
			}
			err = commands.SecretCreate(apiClient, secretText)
		}

	default:
		slog.Error("Unknown command", slog.String("command", command))
		os.Exit(1)
	}

	if err != nil {
		slog.Error(err.Error())
		os.Exit(1)
	}
}

func handleConfigCommand() error {
	if len(os.Args) < 3 {
		return errors.New("usage: seqre config <set|get|clipboard> [args]")
	}

	subcommand := os.Args[2]
	switch subcommand {
	case "set":
		if len(os.Args) < 4 {
			return errors.New("usage: seqre config set <server>")
		}
		return commands.ConfigSet(os.Args[3], version)

	case "get":
		return commands.ConfigGet()

	case "clipboard":
		if len(os.Args) < 4 {
			return errors.New("usage: seqre config clipboard <on|off>")
		}
		switch os.Args[3] {
		case "on":
			return commands.ConfigEnableClipboard()
		case "off":
			return commands.ConfigDisableClipboard()
		default:
			return fmt.Errorf("invalid option '%s'. Use 'on' or 'off'", os.Args[3])
		}

	default:
		return fmt.Errorf("unknown config subcommand: %s", subcommand)
	}
}

func printUsage() {
	_, _ = fmt.Fprint(os.Stdout, "Usage: seqre <command> [args]\n")
	_, _ = fmt.Fprint(os.Stdout, "Commands:\n")
	_, _ = fmt.Fprint(os.Stdout, "  url <URL>                       Create a shortened URL\n")
	_, _ = fmt.Fprint(os.Stdout, "  expand <short>                  Expand a shortened URL\n")
	_, _ = fmt.Fprint(os.Stdout, "  secret <text>                   Create an encrypted secret\n")
	_, _ = fmt.Fprint(os.Stdout, "  secret get <short> <key>        Retrieve and decrypt a secret\n")
	_, _ = fmt.Fprint(os.Stdout, "  ip                              Get your IP address\n")
	_, _ = fmt.Fprint(os.Stdout, "  config set <server>             Set the server URL\n")
	_, _ = fmt.Fprint(os.Stdout, "  config get                      Get the server URL\n")
	_, _ = fmt.Fprint(os.Stdout, "  config clipboard <on|off>       Enable/disable auto-copy to clipboard\n")
	_, _ = fmt.Fprint(os.Stdout, "  version                         Show version information\n")
}
