package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"strconv"
	"todo-app/logging"
	"todo-app/storage"
)

const (
	datafolder string = "tododata"
	datafile   string = "todos.json"
	logfile    string = "todos.log"
)

type runmode int

const (
	RunModeCLI int = iota
	RunModeServer
)

// var ServerLogger logging.AppLogger
var runMode runmode

func main() {
	// input flags
	var flagCreate = flag.String("create", "", "create todo task item (\"description\")")
	var flagUpdate = flag.Int("update", 0, "update todo task item description (id -description \"new description\")")
	var flagNotStarted = flag.Int("not_started", 0, "set todo task item status to not started ( id )")
	var flagStarted = flag.Int("start", 0, "set todo task item status to started ( id )")
	var flagCompleted = flag.Int("complete", 0, "set todo task item status to completed ( id )")
	var flagDelete = flag.Int("delete", 0, "delete a todo task item ( id )")
	var flagList = flag.Bool("list", false, "list items in the todo list ( optionally use -itemid num to show one item)")

	// item description for create and update
	var itemDescription string
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
	})

	// // grab the flag input state from command line
	flag.Parse()

	runMode = runmode(RunModeCLI)

	// setup application context with trace id
	id := logging.GenerateID()
	ctx := context.WithValue(context.Background(), logging.TraceIDKey, id)

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
		logging.Setup(logFileHandle, logOptions)
		logging.Log().InfoContext(ctx, "Starting up logging with static logger")
	}

	// init / pickup current list before process command
	storagefile := fmt.Sprintf("%s\\%s", dir, datafile)
	// open the data file for cli and api
	openErr := storage.Open(ctx, storagefile)
	if openErr != nil {
		// log file not ready so default std.err logging here
		logging.Log().ErrorContext(ctx, "Open file failed, cannot continue", "error", openErr, "datafile", storagefile)
		fmt.Printf("Open file failed, cannot continue,"+" error: %s, datafile: %s\n", openErr, storagefile)
		return
	}

	// process the flags
	switch {
	case *flagCreate != "":
		if nextItem, ok := storage.CreateItem(ctx, *flagCreate, itemDescription); ok == nil {
			storage.ListItem(nextItem)
		}
	case *flagUpdate > 0 && len(itemDescription) > 0:
		if ok := storage.UpdateDescription(ctx, *flagUpdate, itemDescription); ok == nil {
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
	case *flagList:
		storage.ListItem(itemID)
	}

	if runMode == runmode(RunModeCLI) {
		// write back to the file
		storage.Save(ctx, storagefile)
	}
}
