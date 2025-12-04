package factory

import (
	"github.com/spf13/afero"

	"github.com/gizzahub/gzh-cli-shellforge/internal/app"
	"github.com/gizzahub/gzh-cli-shellforge/internal/domain"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/filesystem"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/git"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/snapshot"
	"github.com/gizzahub/gzh-cli-shellforge/internal/infra/yamlparser"
)

// Services holds all common services used by CLI commands
type Services struct {
	Fs     afero.Fs
	Parser *yamlparser.Parser
	Reader *filesystem.Reader
	Writer *filesystem.Writer
}

// NewServices creates a new Services instance with all common dependencies
func NewServices() *Services {
	fs := afero.NewOsFs()
	return &Services{
		Fs:     fs,
		Parser: yamlparser.New(fs),
		Reader: filesystem.NewReader(fs),
		Writer: filesystem.NewWriter(fs),
	}
}

// NewBuilder creates a BuilderService from the services
func (s *Services) NewBuilder() *app.BuilderService {
	return app.NewBuilderService(s.Parser, s.Reader, s.Writer)
}

// BackupServices holds services specifically for backup operations
type BackupServices struct {
	Fs            afero.Fs
	Config        *domain.BackupConfig
	SnapshotMgr   *snapshot.Manager
	GitRepo       app.GitRepository
	BackupService *app.BackupService
}

// BackupOptions configures the backup services
type BackupOptions struct {
	BackupDir  string
	GitEnabled bool
	KeepCount  int
	KeepDays   int
}

// NewBackupServices creates all services needed for backup/restore/cleanup operations
func NewBackupServices(opts BackupOptions) *BackupServices {
	fs := afero.NewOsFs()
	config := domain.NewBackupConfig(opts.BackupDir)
	config.GitEnabled = opts.GitEnabled

	if opts.KeepCount > 0 {
		config.KeepCount = opts.KeepCount
	}
	if opts.KeepDays > 0 {
		config.KeepDays = opts.KeepDays
	}

	snapshotMgr := snapshot.NewManager(fs, config)
	gitRepo := newGitRepositoryAdapter(git.NewRepository(opts.BackupDir))
	backupService := app.NewBackupService(snapshotMgr, gitRepo, config)

	return &BackupServices{
		Fs:            fs,
		Config:        config,
		SnapshotMgr:   snapshotMgr,
		GitRepo:       gitRepo,
		BackupService: backupService,
	}
}

// GitRepositoryAdapter adapts git.Repository to app.GitRepository interface
type GitRepositoryAdapter struct {
	repo *git.Repository
}

// NewGitRepositoryAdapter creates a new GitRepositoryAdapter
func NewGitRepositoryAdapter(repo *git.Repository) *GitRepositoryAdapter {
	return &GitRepositoryAdapter{repo: repo}
}

func newGitRepositoryAdapter(repo *git.Repository) *GitRepositoryAdapter {
	return NewGitRepositoryAdapter(repo)
}

func (a *GitRepositoryAdapter) IsGitInstalled() bool {
	return git.IsGitInstalled()
}

func (a *GitRepositoryAdapter) Init() error {
	return a.repo.Init()
}

func (a *GitRepositoryAdapter) IsInitialized() bool {
	return a.repo.IsInitialized()
}

func (a *GitRepositoryAdapter) ConfigUser(name, email string) error {
	return a.repo.ConfigUser(name, email)
}

func (a *GitRepositoryAdapter) AddAndCommit(message string, paths ...string) error {
	return a.repo.AddAndCommit(message, paths...)
}

func (a *GitRepositoryAdapter) HasChanges() (bool, error) {
	return a.repo.HasChanges()
}
