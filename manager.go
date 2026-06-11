// Package motoig provides an unofficial Instagram client library.
package motoig

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"sync"

	igerr "github.com/motovax/motoig/errors"
	"github.com/motovax/motoig/state"
	"github.com/motovax/motoig/storage"
)

// AccountSpec describes one managed Instagram account.
type AccountSpec struct {
	ID        string `json:"id"`
	Username  string `json:"username,omitempty"`
	Password  string `json:"password,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
	ProxyURL  string `json:"proxy,omitempty"`
}

func (a AccountSpec) options() []Option {
	var opts []Option
	if a.UserAgent != "" {
		opts = append(opts, WithUserAgent(a.UserAgent))
	}
	if a.ProxyURL != "" {
		opts = append(opts, WithProxy(a.ProxyURL))
	}
	return opts
}

type accountsFile struct {
	Accounts []AccountSpec `json:"accounts"`
}

// LoadAccountSpecs reads account definitions from a JSON file.
func LoadAccountSpecs(path string) ([]AccountSpec, error) {
	b, err := os.ReadFile(path)
	if err != nil {
		return nil, igerr.Wrap("LoadAccountSpecs", "read file", err)
	}
	var file accountsFile
	if err := json.Unmarshal(b, &file); err != nil {
		return nil, igerr.Wrap("LoadAccountSpecs", "parse json", err)
	}
	if len(file.Accounts) == 0 {
		return nil, igerr.New("LoadAccountSpecs", "no accounts defined")
	}
	for i, a := range file.Accounts {
		if a.ID == "" {
			return nil, igerr.New("LoadAccountSpecs", fmt.Sprintf("accounts[%d]: missing id", i))
		}
	}
	return file.Accounts, nil
}

// Manager orchestrates multiple Client instances with isolated sessions.
type Manager struct {
	Clients map[string]*Client
	storage storage.Store
	log     *slog.Logger

	mu sync.RWMutex
}

// NewManager creates a multi-account manager.
func NewManager(store storage.Store, log *slog.Logger) *Manager {
	if log == nil {
		log = slog.Default()
	}
	return &Manager{
		Clients: make(map[string]*Client),
		storage: store,
		log:     log,
	}
}

// NewManagerWithSQLite uses SQLite session storage.
func NewManagerWithSQLite(dbPath string, log *slog.Logger) (*Manager, error) {
	store, err := storage.OpenSQLite(dbPath)
	if err != nil {
		return nil, err
	}
	return NewManager(store, log), nil
}

// NewManagerWithDir uses JSON file storage under dir.
func NewManagerWithDir(dir string, log *slog.Logger) *Manager {
	return NewManager(&storage.JSONStore{Directory: dir}, log)
}

// NewManagerWithJSONFile uses a single JSON file for all sessions.
func NewManagerWithJSONFile(path string, log *slog.Logger) *Manager {
	return NewManager(&storage.MultiJSONStore{Path: path}, log)
}

// ClientIDs returns registered account ids.
func (m *Manager) ClientIDs() []string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	ids := make([]string, 0, len(m.Clients))
	for id := range m.Clients {
		ids = append(ids, id)
	}
	return ids
}

// GetClient returns a managed client by id.
func (m *Manager) GetClient(clientID string) (*Client, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	c := m.Clients[clientID]
	if c == nil {
		return nil, igerr.New("GetClient", "unknown client id: "+clientID)
	}
	return c, nil
}

// ImportSettings saves session settings to storage for clientID.
func (m *Manager) ImportSettings(ctx context.Context, clientID string, settings map[string]any) error {
	if m.storage == nil {
		return igerr.New("ImportSettings", "no session storage configured")
	}
	return m.storage.Save(ctx, clientID, settings)
}

// RestoreClient loads settings from storage and registers the account.
func (m *Manager) RestoreClient(ctx context.Context, clientID string, opts ...Option) (*Client, error) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, ok := m.Clients[clientID]; ok {
		return nil, igerr.New("RestoreClient", "client id already registered: "+clientID)
	}

	cfg := clientConfig{log: m.log}
	for _, o := range opts {
		o(&cfg)
	}

	st, err := m.stateFromStorage(ctx, clientID, cfg)
	if err != nil {
		return nil, err
	}

	c := newClient(st, cfg)
	m.Clients[clientID] = c
	return c, nil
}

func (m *Manager) stateFromStorage(ctx context.Context, clientID string, cfg clientConfig) (*state.State, error) {
	if m.storage == nil {
		return nil, igerr.New("RestoreClient", "no session storage configured")
	}
	snap, err := m.storage.Load(ctx, clientID)
	if err != nil {
		return nil, err
	}
	if snap == nil {
		return nil, igerr.New("RestoreClient", "no session in storage for client id: "+clientID)
	}

	s := state.New(state.Options{
		UserAgent: cfg.userAgent,
		ProxyURL:  cfg.proxyURL,
	})
	s.LoadSnapshot(snap)
	return s, nil
}

// AddAccounts registers multiple accounts.
func (m *Manager) AddAccounts(ctx context.Context, accounts ...AccountSpec) error {
	var errs []error
	for _, spec := range accounts {
		c, err := m.RestoreClient(ctx, spec.ID, spec.options()...)
		if err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", spec.ID, err))
			continue
		}
		m.log.Info("account registered", "client_id", spec.ID, "uid", c.UserID(), "username", c.Username())
	}
	return errors.Join(errs...)
}

// AddAccountsFromFile loads account specs from JSON and registers them.
func (m *Manager) AddAccountsFromFile(ctx context.Context, path string) error {
	specs, err := LoadAccountSpecs(path)
	if err != nil {
		return err
	}
	return m.AddAccounts(ctx, specs...)
}

// StoredClientIDs returns client ids with persisted sessions.
func (m *Manager) StoredClientIDs(ctx context.Context) ([]string, error) {
	if m.storage == nil {
		return nil, nil
	}
	lister, ok := m.storage.(storage.Lister)
	if !ok {
		return nil, igerr.New("StoredClientIDs", "storage does not support listing")
	}
	return lister.List(ctx)
}

// RestoreAll registers every client id found in storage.
func (m *Manager) RestoreAll(ctx context.Context) error {
	ids, err := m.StoredClientIDs(ctx)
	if err != nil {
		return err
	}
	var errs []error
	for _, id := range ids {
		m.mu.RLock()
		_, exists := m.Clients[id]
		m.mu.RUnlock()
		if exists {
			continue
		}
		if _, err := m.RestoreClient(ctx, id); err != nil {
			errs = append(errs, fmt.Errorf("%s: %w", id, err))
		}
	}
	return errors.Join(errs...)
}

// SaveSession persists one client session.
func (m *Manager) SaveSession(ctx context.Context, clientID string) error {
	if m.storage == nil {
		return igerr.New("SaveSession", "no session storage configured")
	}
	m.mu.RLock()
	c := m.Clients[clientID]
	m.mu.RUnlock()
	if c == nil {
		return igerr.New("SaveSession", "no active session for client id: "+clientID)
	}
	return m.storage.Save(ctx, clientID, c.GetSettings())
}

// SaveAllSessions persists every active client session.
func (m *Manager) SaveAllSessions(ctx context.Context) error {
	if m.storage == nil {
		return igerr.New("SaveAllSessions", "no session storage configured")
	}
	var errs []error
	for _, id := range m.ClientIDs() {
		if err := m.SaveSession(ctx, id); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// RemoveClient removes a client.
func (m *Manager) RemoveClient(ctx context.Context, clientID string, persist bool) error {
	m.mu.Lock()
	c := m.Clients[clientID]
	delete(m.Clients, clientID)
	m.mu.Unlock()
	if c == nil {
		return nil
	}
	if persist && m.storage != nil {
		_ = m.storage.Save(ctx, clientID, c.GetSettings())
	}
	return c.State().Close()
}

// Close stops all clients and optionally persists sessions.
func (m *Manager) Close(ctx context.Context, persist bool) error {
	if persist && m.storage != nil {
		_ = m.SaveAllSessions(ctx)
	}
	var errs []error
	for _, id := range m.ClientIDs() {
		if err := m.RemoveClient(ctx, id, false); err != nil {
			errs = append(errs, err)
		}
	}
	if closer, ok := m.storage.(interface{ Close() error }); ok {
		if err := closer.Close(); err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}
