package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/hacobu-inc/mcp-goref/internal/refactor"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	// Configuration flags
	// MCP server name (client-visible identifier)
	version := flag.String("version", "0.1.0", "MCP server version")
	flag.Parse()

	// Create MCP server with name, version, and instructions
	srv := server.NewMCPServer(
		"go-refactor-mcp-server",
		*version,
		// Instructions for clients: describe server purpose and available tools
		server.WithInstructions(
			strings.Join([]string{
				"This is an MCP server for refactoring in Go language.",
				"This server provides tools to list symbols (such as structs and functions) within a Go module and rename them.",
				"The following tools are available:",
				"- list-refs: Search for all references to a specified symbol and obtain a list.",
				"- rename: Rename a specified symbol.",
			}, "\n"),
		),
		server.WithLogging(),
		server.WithRecovery(),
	)

	// Tool: list-refs
	// Lists all occurrences of a given symbol within the module containing the specified Go source file.
	listRefs := mcp.NewTool("list-refs",
		mcp.WithDescription("Lists all occurrences of a given symbol within the module containing the specified Go source file."),
		mcp.WithString("file",
			mcp.Required(),
			mcp.Description("The path to the Go source file containing the definition of the symbol you want to specify. This is used to identify the package to which the symbol belongs."),
		),
		mcp.WithString("symbol",
			mcp.Required(),
			mcp.Description("The name of the symbol you want to obtain references for. You can specify functions, variables, types, structs, or methods (Type.Method)."),
		),
	)
	srv.AddTool(listRefs, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		// capture stdout from ListRefs
		fileArg, _ := req.Params.Arguments["file"].(string)
		sym, _ := req.Params.Arguments["symbol"].(string)
		oldOut := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		os.Stdout = w
		err = refactor.ListRefs(fileArg, sym)
		w.Close()
		os.Stdout = oldOut
		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(buf.String()), nil
	})

	// Tool: rename
	// Renames a symbol throughout the Go module containing the specified file.
	// Symbol may be a function, variable, type, or method (formatted as Type.Method).
	renameTool := mcp.NewTool("rename",
		mcp.WithDescription("Renames a symbol throughout the Go module containing the specified file."),
		mcp.WithString("file",
			mcp.Required(),
			mcp.Description("The path to the Go source file containing the definition of the symbol you want to rename. This is used to identify the package to which the symbol belongs."),
		),
		mcp.WithString("oldSymbol",
			mcp.Required(),
			mcp.Description("The name of the symbol you want to rename. You can specify functions, variables, types, structs, or methods (Type.Method)."),
		),
		mcp.WithString("newSymbol",
			mcp.Required(),
			mcp.Description("The name of the symbol after renaming. You can specify functions, variables, types, structs, or methods (Type.Method)."),
		),
	)
	srv.AddTool(renameTool, func(ctx context.Context, req mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		fileArg, _ := req.Params.Arguments["file"].(string)
		oldName, _ := req.Params.Arguments["oldSymbol"].(string)
		newName, _ := req.Params.Arguments["newSymbol"].(string)
		// capture stdout from Rename
		oldOut := os.Stdout
		r, w, err := os.Pipe()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		os.Stdout = w
		err = refactor.Rename(fileArg, oldName, newName)
		w.Close()
		os.Stdout = oldOut
		var buf bytes.Buffer
		io.Copy(&buf, r)
		r.Close()
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}
		return mcp.NewToolResultText(buf.String()), nil
	})

	// Start server over stdio
	if err := server.ServeStdio(srv); err != nil {
		fmt.Fprintf(os.Stderr, "MCP server error: %v\n", err)
		os.Exit(1)
	}
}
