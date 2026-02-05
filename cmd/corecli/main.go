package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"strings"

	"taskpp/core/bind"
)

type cmdConfig struct {
	DBPath string
	Pass   string
	Init   bool
}

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(2)
	}

	cfg := cmdConfig{}
	rootFlags := flag.NewFlagSet("corecli", flag.ExitOnError)
	rootFlags.StringVar(&cfg.DBPath, "db", "taskpp.db", "path to sqlite db")
	rootFlags.StringVar(&cfg.Pass, "pass", "", "passphrase (auto-unlock)")
	rootFlags.BoolVar(&cfg.Init, "init", false, "initialize keys if needed")
	_ = rootFlags.Parse(os.Args[1:])

	args := rootFlags.Args()
	if len(args) == 0 {
		printUsage()
		os.Exit(2)
	}

	core := newCore(cfg)
	if errStr := core.Open(); errStr != "" {
		fatal(errStr)
	}
	defer core.Close()

	if strings.TrimSpace(cfg.Pass) != "" {
		if cfg.Init {
			if errStr := core.InitKeys(cfg.Pass); errStr != "" && !isAlreadyInit(errStr) {
				fatal(errStr)
			}
		}
		if errStr := core.UnlockKeys(cfg.Pass); errStr != "" {
			fatal(errStr)
		}
	}

	switch args[0] {
	case "init-keys":
		cmdInitKeys(core, args[1:])
	case "unlock-keys":
		cmdUnlockKeys(core, args[1:])
	case "add":
		cmdAdd(core, args[1:])
	case "list":
		cmdList(core, args[1:])
	case "update":
		cmdUpdate(core, args[1:])
	case "done":
		cmdDone(core, args[1:])
	case "due":
		cmdDue(core, args[1:])
	case "reorder":
		cmdReorder(core, args[1:])
	case "export":
		cmdExport(core, args[1:])
	case "import":
		cmdImport(core, args[1:])
	case "decrypt-event":
		cmdDecryptEvent(core, args[1:])
	case "delete":
		cmdDelete(core, args[1:])
	default:
		printUsage()
		os.Exit(2)
	}
}

func newCore(cfg cmdConfig) *bind.Core {
	config := bind.Config{
		StorageDriver: "sqlite",
		StoragePath:   "file:" + cfg.DBPath,
	}
	data, _ := json.Marshal(config)
	return bind.NewCore(string(data))
}

func cmdAdd(core *bind.Core, args []string) {
	fs := flag.NewFlagSet("add", flag.ExitOnError)
	title := fs.String("title", "", "task title")
	desc := fs.String("desc", "", "task description")
	priority := fs.String("priority", "", "low|med|high")
	due := fs.String("due", "", "YYYY-MM-DD")
	_ = fs.Parse(args)

	if strings.TrimSpace(*title) == "" {
		fatal("title is required")
	}

	dto := bind.TaskDTO{
		Title:       *title,
		Description: *desc,
		Priority:    *priority,
		DueDate:     *due,
	}
	payload, _ := json.Marshal(dto)
	result := core.CreateTask(string(payload))
	printJSON(result)
}

func cmdInitKeys(core *bind.Core, args []string) {
	fs := flag.NewFlagSet("init-keys", flag.ExitOnError)
	pass := fs.String("pass", "", "passphrase")
	_ = fs.Parse(args)
	if strings.TrimSpace(*pass) == "" {
		fatal("pass is required")
	}
	result := core.InitKeys(*pass)
	printJSON(result)
}

func cmdUnlockKeys(core *bind.Core, args []string) {
	fs := flag.NewFlagSet("unlock-keys", flag.ExitOnError)
	pass := fs.String("pass", "", "passphrase")
	_ = fs.Parse(args)
	if strings.TrimSpace(*pass) == "" {
		fatal("pass is required")
	}
	result := core.UnlockKeys(*pass)
	printJSON(result)
}

func cmdList(core *bind.Core, args []string) {
	fs := flag.NewFlagSet("list", flag.ExitOnError)
	status := fs.String("status", "", "active|done")
	archived := fs.String("archived", "", "true|false")
	due := fs.String("due", "", "YYYY-MM-DD")
	_ = fs.Parse(args)

	var archivedPtr *bool
	if *archived != "" {
		value := *archived == "true"
		archivedPtr = &value
	}

	filter := bind.TaskFilterDTO{
		Status:   *status,
		Archived: archivedPtr,
		DueDate:  *due,
	}
	payload, _ := json.Marshal(filter)
	result := core.ListTasks(string(payload))
	printJSON(result)
}

func cmdUpdate(core *bind.Core, args []string) {
	fs := flag.NewFlagSet("update", flag.ExitOnError)
	id := fs.String("id", "", "task id")
	title := fs.String("title", "", "task title")
	desc := fs.String("desc", "", "task description")
	status := fs.String("status", "", "active|done")
	priority := fs.String("priority", "", "low|med|high")
	due := fs.String("due", "", "YYYY-MM-DD")
	archived := fs.String("archived", "", "true|false")
	_ = fs.Parse(args)

	if strings.TrimSpace(*id) == "" {
		fatal("id is required")
	}

	dto, err := loadTask(core, *id)
	if err != nil {
		fatal(err.Error())
	}
	if *title != "" {
		dto.Title = *title
	}
	if *desc != "" {
		dto.Description = *desc
	}
	if *status != "" {
		dto.Status = *status
	}
	if *priority != "" {
		dto.Priority = *priority
	}
	if *due != "" {
		dto.DueDate = *due
	}
	if *archived != "" {
		dto.Archived = *archived == "true"
	}
	payload, _ := json.Marshal(dto)
	result := core.UpdateTask(string(payload))
	printJSON(result)
}

func cmdDone(core *bind.Core, args []string) {
	if len(args) < 1 {
		fatal("usage: done <task-id>")
	}
	result := core.SetCompleted(args[0], true)
	printJSON(result)
}

func cmdDue(core *bind.Core, args []string) {
	if len(args) < 2 {
		fatal("usage: due <task-id> <YYYY-MM-DD>")
	}
	result := core.SetDueDate(args[0], args[1])
	printJSON(result)
}

func cmdReorder(core *bind.Core, args []string) {
	fs := flag.NewFlagSet("reorder", flag.ExitOnError)
	items := fs.String("items", "", "comma-separated id:order:due_date")
	_ = fs.Parse(args)

	if strings.TrimSpace(*items) == "" {
		fatal("items is required")
	}

	parts := strings.Split(*items, ",")
	payloadItems := make([]bind.ReorderItemDTO, 0, len(parts))
	for _, part := range parts {
		segment := strings.Split(part, ":")
		if len(segment) < 2 {
			fatal("invalid items format")
		}
		order, err := parseInt64(segment[1])
		if err != nil {
			fatal("invalid order")
		}
		item := bind.ReorderItemDTO{
			ID:    segment[0],
			Order: order,
		}
		if len(segment) >= 3 {
			item.DueDate = segment[2]
		}
		payloadItems = append(payloadItems, item)
	}
	payload, _ := json.Marshal(payloadItems)
	result := core.ReorderTasks(string(payload))
	printJSON(result)
}

func cmdExport(core *bind.Core, args []string) {
	fs := flag.NewFlagSet("export", flag.ExitOnError)
	since := fs.Int64("since", 0, "last seq")
	_ = fs.Parse(args)
	result := core.ExportEvents(*since)
	printJSON(result)
}

func cmdImport(core *bind.Core, args []string) {
	fs := flag.NewFlagSet("import", flag.ExitOnError)
	payload := fs.String("events", "", "events json")
	_ = fs.Parse(args)
	if strings.TrimSpace(*payload) == "" {
		fatal("events is required")
	}
	result := core.ImportEvents(*payload)
	printJSON(result)
}

func cmdDecryptEvent(core *bind.Core, args []string) {
	fs := flag.NewFlagSet("decrypt-event", flag.ExitOnError)
	payload := fs.String("payload", "", "base64 payload")
	_ = fs.Parse(args)
	if strings.TrimSpace(*payload) == "" {
		fatal("payload is required")
	}
	result := core.DebugDecryptEvent(*payload)
	printJSON(result)
}

func cmdDelete(core *bind.Core, args []string) {
	if len(args) < 1 {
		fatal("usage: delete <task-id>")
	}
	result := core.DeleteTask(args[0])
	printJSON(result)
}

func printJSON(payload string) {
	if payload == "" {
		fmt.Println("ok")
		return
	}
	var out any
	if err := json.Unmarshal([]byte(payload), &out); err != nil {
		fmt.Println(payload)
		return
	}
	pretty, err := json.MarshalIndent(out, "", "  ")
	if err != nil {
		fmt.Println(payload)
		return
	}
	fmt.Println(string(pretty))
}

func fatal(message string) {
	fmt.Fprintln(os.Stderr, message)
	os.Exit(1)
}

func printUsage() {
	fmt.Println("corecli -db <path> [-pass <passphrase>] [-init] <command> [args]")
	fmt.Println("commands:")
	fmt.Println("  init-keys -pass <passphrase>")
	fmt.Println("  unlock-keys -pass <passphrase>")
	fmt.Println("  add    -title <t> [-desc <d>] [-priority low|med|high] [-due YYYY-MM-DD]")
	fmt.Println("  list   [-status active|done] [-archived true|false] [-due YYYY-MM-DD]")
	fmt.Println("  update -id <id> [-title <t>] [-desc <d>] [-status active|done] [-priority low|med|high] [-due YYYY-MM-DD] [-archived true|false]")
	fmt.Println("  done   <task-id>")
	fmt.Println("  due    <task-id> <YYYY-MM-DD>")
	fmt.Println("  reorder -items id:order[:due_date],id:order[:due_date]")
	fmt.Println("  export [-since <seq>]")
	fmt.Println("  import -events <json>")
	fmt.Println("  decrypt-event -payload <base64>")
	fmt.Println("  delete <task-id>")
}

func parseInt64(input string) (int64, error) {
	var value int64
	_, err := fmt.Sscanf(input, "%d", &value)
	return value, err
}

func isAlreadyInit(errStr string) bool {
	return strings.Contains(errStr, "keys already initialized")
}

func loadTask(core *bind.Core, id string) (bind.TaskDTO, error) {
	result := core.ListTasks("")
	var tasks []bind.TaskDTO
	if err := json.Unmarshal([]byte(result), &tasks); err != nil {
		return bind.TaskDTO{}, fmt.Errorf("decode tasks: %v", err)
	}
	for _, task := range tasks {
		if task.ID == id {
			return task, nil
		}
	}
	return bind.TaskDTO{}, fmt.Errorf("task not found: %s", id)
}
