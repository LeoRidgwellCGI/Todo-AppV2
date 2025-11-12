package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"todo-app/logging"
	"todo-app/storage"
)

const (
	datafolder string = "tododata"
	datafile   string = "todos.json"
	logfile    string = "todos.log"
)

type RunMode string

const (
	RunModeCLI    = "CLI"
	RunModeServer = "SERVER"
)

type ctxKey string

const traceIDKey ctxKey = "Trace ID"

type ContextHandler struct {
	slog.Handler
}

var (
	runMode RunMode
)

// Handle adds context information (like Trace ID) to the log record before passing it to the underlying handler.
func (h *ContextHandler) Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		r.AddAttrs(slog.String(string(traceIDKey), traceID))
	}
	return h.Handler.Handle(ctx, r)
}

func main() {
	// default to cli mode
	runMode = RunModeCLI

	// input flags
	var flagCreate = flag.String("create", "", "create todo task item (\"description\") (optionally use -status \"not_started|has_started|completed\")")
	var flagUpdate = flag.Int("update", 0, "update todo task item description (id -description \"new description\") (optionally use -status \"not_started|has_started|completed\")")
	var flagDelete = flag.Int("delete", 0, "delete a todo task item ( id )")
	var flagList = flag.Bool("list", false, "list items in the todo list ( optionally use -itemid num to show one item)")
	var flagStatus = flag.String("status", "", "use this with -create or -update to set the status (\"not_started|has_started|completed\")")
	var flagDescription = flag.String("description", "", "use this with -update for the update description text -description \"new text\"")
	var flagItemID = flag.Int("itemid", 0, "optional, use this -itemid with -list for one item")
	flag.Parse()

	// setup application context with trace id
	traceID := logging.GenerateID()
	ctx := context.WithValue(context.Background(), traceIDKey, traceID)

	// resolve the appdata data sub folder
	dir, err := logging.CreateAppDataFolder(datafolder)
	if err != nil {
		// don't have a file logger yet!
		fmt.Printf("Cannot establish working data folder")
		return
	}

	// wire up logger
	logName := dir + "\\" + logfile
	if logFileHandle, err := logging.OpenLogFile(logName); err == nil {
		defer logFileHandle.Close()
		logOptions := logging.LoggerOptions()
		slog.SetDefault(slog.New(&ContextHandler{slog.NewTextHandler(logFileHandle, &logOptions)}))
		slog.InfoContext(ctx, "Starting up logging with static logger")
	}

	// init / pickup current list before process command
	storagefile := fmt.Sprintf("%s\\%s", dir, datafile)

	// open the data file for cli and api
	openErr := storage.Open(ctx, storagefile)
	if openErr != nil {
		// log file not ready so default std.err logging here
		slog.ErrorContext(ctx, "Open file failed, cannot continue", "error", openErr, "datafile", storagefile)
		fmt.Printf("Open file failed, cannot continue,"+" error: %s, datafile: %s\n", openErr, storagefile)
		return
	}

	// process the flags
	switch {
	case *flagList:
		storage.ListItem(*flagItemID)
	case *flagCreate != "":
		if *flagStatus != "" {
			if *flagStatus == "not_started" || *flagStatus == "has_started" || *flagStatus == "completed" {
				// valid status
			} else {
				fmt.Fprintf(os.Stderr, "Invalid status value: %s. Use 'not_started', 'has_started', or 'completed'.\n", *flagStatus)
				slog.ErrorContext(ctx, "Invalid status value for create", "Status", *flagStatus)
				*flagStatus = "not_started"
			}
		}
		if newItem, ok := storage.CreateItem(ctx, *flagCreate, *flagStatus); ok == nil {
			storage.ListItem(newItem.ID)
		} else {
			fmt.Fprintf(os.Stderr, "Failed to create item.\n")
			slog.ErrorContext(ctx, "Failed to create item", "Description", *flagCreate, "Status", *flagStatus)
		}
	case *flagUpdate > 0:
		if *flagDescription == "" {
			fmt.Fprintf(os.Stderr, "Update requires -description \"new description\" to be set.\n")
			slog.ErrorContext(ctx, "Update missing description", "ItemID", *flagUpdate)
			break
		}

		// get existing item
		if item, ok := storage.GetItemByID(*flagUpdate); ok == nil {
			newItem := item
			newItem.Description = *flagDescription
			if *flagStatus == "not_started" || *flagStatus == "has_started" || *flagStatus == "completed" {
				newItem.Status = *flagStatus
			} else {
				fmt.Fprintf(os.Stderr, "Invalid status value: %s. Use 'not_started', 'has_started', or 'completed'.\n", *flagStatus)
				slog.ErrorContext(ctx, "Invalid status value for update", "Status", *flagStatus)
				break
			}

			// perform the update
			if _, ok := storage.UpdateItem(ctx, newItem); ok == nil {
				storage.ListItem(*flagUpdate)
			} else {
				fmt.Fprintf(os.Stderr, "Failed to update item ID %d.\n", *flagUpdate)
				slog.ErrorContext(ctx, "Failed to update item", "ItemID", *flagUpdate)
			}
		} else {
			fmt.Fprintf(os.Stderr, "Item ID %d not found for update.\n", *flagUpdate)
			slog.ErrorContext(ctx, "Item ID not found for update", "ItemID", *flagUpdate)
		}
	case *flagDelete > 0:
		if ok := storage.DeleteItem(ctx, *flagDelete); ok == nil {
			storage.ListItem(0)
		} else {
			fmt.Fprintf(os.Stderr, "Item ID %d not found for delete.\n", *flagDelete)
			slog.ErrorContext(ctx, "Item ID not found for delete", "ItemID", *flagDelete)
		}
	default:
		fmt.Fprintf(os.Stderr, `Todo-App
Manage to-do items: list, add, update descriptions, or delete by ID.

Usage:
  go run . -list [-itemid <id>]
  go run . -create "<description> " [-status "not_started|has_started|completed"]
  go run . -update <id> "<new description> " [-status "not_started|has_started|completed"]
  go run . -delete <id>
`)
	}

	if runMode == RunModeCLI {
		// write back to the file
		storage.Save(ctx, storagefile)
	}
}
