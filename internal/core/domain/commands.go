package domain

const (
	CommandTypeManageRepo = "manage_repository"
	CommandTypeVerifyRepo = "verify_repository"
)

type RepositoryCommand struct {
	Command
	RepositoryURL  string
	RepositoryName string
}
