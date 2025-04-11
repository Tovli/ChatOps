package domain

// Command represents a platform-agnostic command
type Command struct {
    ID          string
    Type        string
    Parameters  map[string]interface{}
    Source      CommandSource
    User        User
    Timestamp   time.Time
}

type CommandSource struct {
    Platform    string    // "slack", "teams", etc.
    ChannelID   string
    MessageID   string
}

type User struct {
    ID          string
    Platform    string
    Permissions []string
}

type WorkflowTrigger struct {
    Type        string
    Repository  string
    Workflow    string
    Parameters  map[string]interface{}
}

type CommandResult struct {
    Status      string
    Message     string
    Details     interface{}
    Error       error
} 