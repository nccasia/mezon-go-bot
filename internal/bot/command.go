package bot

// TODO: multiple command by ws channel message
type CommandHandler func(command string, args []string) string
