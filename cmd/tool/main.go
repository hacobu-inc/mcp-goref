package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/hacobu-inc/mcp-goref/internal/refactor"
)

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(1)
	}
	cmd := os.Args[1]
	switch cmd {
	case "list-refs":
		if len(os.Args) != 4 {
			fmt.Fprintf(os.Stderr, "Usage: %s list-refs <file.go> <symbolName>\n", filepath.Base(os.Args[0]))
			os.Exit(1)
		}
		fileArg := os.Args[2]
		symbol := os.Args[3]
		if err := refactor.ListRefs(fileArg, symbol); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "rename":
		if len(os.Args) != 5 {
			fmt.Fprintf(os.Stderr, "Usage: %s rename <file.go> <oldSymbolName> <newSymbolName>\n", filepath.Base(os.Args[0]))
			os.Exit(1)
		}
		fileArg := os.Args[2]
		oldName := os.Args[3]
		newName := os.Args[4]
		if err := refactor.Rename(fileArg, oldName, newName); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	case "help":
		prog := filepath.Base(os.Args[0])
		switch len(os.Args) {
		case 2:
			usage()
			os.Exit(0)
		case 3:
			topic := os.Args[2]
			switch topic {
			case "list-refs":
				fmt.Fprintf(os.Stderr, "Usage: %s list-refs <file.go> <symbolName>\n", prog)
			case "rename":
				fmt.Fprintf(os.Stderr, "Usage: %s rename <file.go> <oldSymbolName> <newSymbolName>\n", prog)
			case "help":
				fmt.Fprintf(os.Stderr, "Usage: %s help [<command>]\n", prog)
			default:
				fmt.Fprintf(os.Stderr, "%s: unknown help topic %q\n", prog, topic)
				usage()
				os.Exit(1)
			}
			os.Exit(0)
		default:
			fmt.Fprintf(os.Stderr, "Usage: %s help [<command>]\n", prog)
			os.Exit(1)
		}
	default:
		usage()
		os.Exit(1)
	}
}

func usage() {
	prog := filepath.Base(os.Args[0])
	fmt.Fprintf(os.Stderr, "Usage: %s <command> [<args>]\n", prog)
	fmt.Fprintln(os.Stderr, "Commands:")
	fmt.Fprintln(os.Stderr, "  list-refs <file.go> <symbolName>")
	fmt.Fprintln(os.Stderr, "  rename    <file.go> <oldSymbolName> <newSymbolName>")
	fmt.Fprintln(os.Stderr, "  help      [<command>]       Show help for commands")
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintf(os.Stderr, "Use \"%s help <command>\" for more information about a command.\n", prog)
}
