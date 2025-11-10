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

const traceIDKey ctxKey = "trace_id"

type ContextHandler struct {
	slog.Handler
}

var (
	runMode RunMode
	handler ContextHandler
)

// Handle adds context information (like trace_id) to the log record before passing it to the underlying handler.
func Handle(ctx context.Context, r slog.Record) error {
	if traceID, ok := ctx.Value(traceIDKey).(string); ok {
		r.AddAttrs(slog.String(string(traceIDKey), traceID))
	}
	return handler.Handle(ctx, r)
}

func main() {
	// default to cli mode
	runMode = RunModeCLI
	// input flags
	var flagCreate = flag.String("create", "", "create todo task item (\"description\")")
	var flagUpdate = flag.Int("update", 0, "update todo task item description (id -description \"new description\")")
	var flagNotStarted = flag.Int("not_started", 0, "set todo task item status to not started ( id )")
	var flagStarted = flag.Int("started", 0, "set todo task item status to started ( id )")
	var flagCompleted = flag.Int("completed", 0, "set todo task item status to completed ( id )")
	var flagDelete = flag.Int("delete", 0, "delete a todo task item ( id )")
	var flagList = flag.Bool("list", false, "list items in the todo list ( optionally use -itemid num to show one item)")
	var flagDescription = flag.String("description", "", "use this with -update for the update description text -description \"new text\"")
	var flagItemID = flag.Int("itemid", 0, "optional, use this -itemid with -list for one item")
	flag.Parse()

	// item description for create and update
	/*var itemDescription string
	flag.Func("description", "use this with -update for the update description text -description \"new text\"", func(s string) error {
		if len(s) == 0 {
			return errors.New("value of description needs to be supplied")
		} else {
			itemDescription = s
		}
		return nil
	})

	// item id for listing one task
	var itemID int
	flag.Func("itemid", "optional, use this -itemid with -list for one item", func(s string) error {
		if i, ok := strconv.Atoi(s); ok != nil {
			return errors.New("value of itemid needs to be supplied")
		} else {
			itemID = i
		}
		return nil
	})*/
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
		if nextItem, ok := storage.CreateItem(ctx, *flagCreate); ok == nil {
			storage.ListItem(nextItem)
		}
	case *flagUpdate > 0 && *flagDescription != "":
		if ok := storage.UpdateDescription(ctx, *flagUpdate, *flagDescription); ok == nil {
			storage.ListItem(*flagUpdate)
		}
	case *flagNotStarted > 0:
		if ok := storage.UpdateStatus(ctx, *flagNotStarted, storage.StatusNotStarted); ok == nil {
			storage.ListItem(*flagNotStarted)
		}
	case *flagStarted > 0:
		if ok := storage.UpdateStatus(ctx, *flagStarted, storage.StatusStarted); ok == nil {
			storage.ListItem(*flagStarted)
		}
	case *flagCompleted > 0:
		if ok := storage.UpdateStatus(ctx, *flagCompleted, storage.StatusCompleted); ok == nil {
			storage.ListItem(*flagCompleted)
		}
	case *flagDelete > 0:
		storage.DeleteItem(ctx, *flagDelete)
		storage.ListItem(-1)
	default:
		fmt.Fprintf(os.Stderr, `Todo-App
Manage to-do items: list, add, update descriptions, or delete by ID.

Usage:
  go run . -list [-itemid <id>]
  go run . -create "<description>"
  go run . -update <id> "<new description>"
  go run . -not_started <id>
  go run . -started <id>
  go run . -completed <id>
  go run . -delete <id>
`)
	}

	if runMode == RunModeCLI {
		// write back to the file
		storage.Save(ctx, storagefile)
	}
}
