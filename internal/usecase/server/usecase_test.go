package server

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"go.uber.org/zap"

	"github.com/xuri/excelize/v2"

	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/dto"
	"github.com/th1enq/ViettelSMS_ServerService/internal/domain/entity"
	domain "github.com/th1enq/ViettelSMS_ServerService/internal/domain/errors"
	repoiface "github.com/th1enq/ViettelSMS_ServerService/internal/domain/repository"
	srv "github.com/th1enq/ViettelSMS_ServerService/internal/domain/service"
	"gorm.io/gorm"
)

// --- Mocks ---

type mockRepo struct {
	existByNameOrIDFn func(ctx context.Context, serverID string, serverName string) (bool, error)
	createFn          func(ctx context.Context, server *entity.Server) error
	deleteFn          func(ctx context.Context, serverID string) error
	getByFieldFn      func(ctx context.Context, field string, value interface{}) (*entity.Server, error)
	updateFn          func(ctx context.Context, server *entity.Server) error
	getServersFn      func(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) ([]*entity.Server, int, error)
	batchCreateFn     func(ctx context.Context, servers []*entity.Server) ([]*string, error)
	updateStatusFn    func(ctx context.Context, serverID string, status entity.ServerStatus) error
}

func (m *mockRepo) ExistByNameOrID(ctx context.Context, serverID string, serverName string) (bool, error) {
	if m.existByNameOrIDFn == nil {
		return false, nil
	}
	return m.existByNameOrIDFn(ctx, serverID, serverName)
}
func (m *mockRepo) Create(ctx context.Context, server *entity.Server) error {
	if m.createFn == nil {
		return nil
	}
	return m.createFn(ctx, server)
}
func (m *mockRepo) Delete(ctx context.Context, serverID string) error {
	if m.deleteFn == nil {
		return nil
	}
	return m.deleteFn(ctx, serverID)
}
func (m *mockRepo) GetByField(ctx context.Context, field string, value interface{}) (*entity.Server, error) {
	if m.getByFieldFn == nil {
		return nil, gorm.ErrRecordNotFound
	}
	return m.getByFieldFn(ctx, field, value)
}
func (m *mockRepo) Update(ctx context.Context, server *entity.Server) error {
	if m.updateFn == nil {
		return nil
	}
	return m.updateFn(ctx, server)
}
func (m *mockRepo) GetServers(ctx context.Context, filter dto.ServerFilterOptions, pagination dto.ServerPaginationOptions) ([]*entity.Server, int, error) {
	if m.getServersFn == nil {
		return nil, 0, nil
	}
	return m.getServersFn(ctx, filter, pagination)
}
func (m *mockRepo) BatchCreate(ctx context.Context, servers []*entity.Server) ([]*string, error) {
	if m.batchCreateFn == nil {
		ids := make([]*string, 0)
		for i := range servers {
			ids = append(ids, &servers[i].ServerID)
		}
		return ids, nil
	}
	return m.batchCreateFn(ctx, servers)
}
func (m *mockRepo) UpdateStatus(ctx context.Context, serverID string, status entity.ServerStatus) error {
	if m.updateStatusFn == nil {
		return nil
	}
	return m.updateStatusFn(ctx, serverID, status)
}

var _ repoiface.ServerRepository = (*mockRepo)(nil)

type mockXLSX struct {
	getRowsFn  func(filePath string) ([][]string, error)
	validateFn func(row []string) error
	parseFn    func(row []string) (*entity.Server, error)
}

func (m *mockXLSX) GetRows(filePath string) ([][]string, error) {
	if m.getRowsFn == nil {
		return nil, nil
	}
	return m.getRowsFn(filePath)
}
func (m *mockXLSX) Validate(row []string) error {
	if m.validateFn == nil {
		return nil
	}
	return m.validateFn(row)
}
func (m *mockXLSX) Parse(row []string) (*entity.Server, error) {
	if m.parseFn == nil {
		return &entity.Server{ServerID: "id", ServerName: "name", IPv4: "1.1.1.1", IntervalTime: 1}, nil
	}
	return m.parseFn(row)
}

var _ srv.XLSXService = (*mockXLSX)(nil)

func newUseCase(r repoiface.ServerRepository, x srv.XLSXService) UseCase {
	return NewServerUseCase(r, x, zap.NewNop())
}

// --- Tests ---

func TestCreateServer_Success(t *testing.T) {
	r := &mockRepo{
		existByNameOrIDFn: func(ctx context.Context, serverID, serverName string) (bool, error) { return false, nil },
		createFn:          func(ctx context.Context, server *entity.Server) error { return nil },
	}
	uc := newUseCase(r, &mockXLSX{})
	req := dto.CreateServerParams{ServerID: "s1", ServerName: "srv", IPv4: "1.1.1.1", IntervalTime: 5}
	got, err := uc.CreateServer(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got.ServerID != "s1" || got.ServerName != "srv" {
		t.Fatalf("unexpected resp: %+v", got)
	}
}

func TestCreateServer_ExistOrRepoErr(t *testing.T) {
	r1 := &mockRepo{existByNameOrIDFn: func(ctx context.Context, id, name string) (bool, error) { return true, nil }}
	uc1 := newUseCase(r1, &mockXLSX{})
	if _, err := uc1.CreateServer(context.Background(), dto.CreateServerParams{ServerID: "a", ServerName: "b", IPv4: "1.1.1.1", IntervalTime: 1}); !errors.Is(err, domain.ErrServerExist) {
		t.Fatalf("want ErrServerExist got %v", err)
	}

	r2 := &mockRepo{existByNameOrIDFn: func(ctx context.Context, id, name string) (bool, error) { return false, fmt.Errorf("boom") }}
	uc2 := newUseCase(r2, &mockXLSX{})
	if _, err := uc2.CreateServer(context.Background(), dto.CreateServerParams{ServerID: "a", ServerName: "b", IPv4: "1.1.1.1", IntervalTime: 1}); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want ErrInternalServer got %v", err)
	}
}

func TestCreateServer_RepoCreateError(t *testing.T) {
	r := &mockRepo{
		existByNameOrIDFn: func(ctx context.Context, id, name string) (bool, error) { return false, nil },
		createFn:          func(ctx context.Context, server *entity.Server) error { return fmt.Errorf("boom") },
	}
	uc := newUseCase(r, &mockXLSX{})
	if _, err := uc.CreateServer(context.Background(), dto.CreateServerParams{ServerID: "s1", ServerName: "srv", IPv4: "1.1.1.1", IntervalTime: 5}); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want internal, got %v", err)
	}
}

func TestDeleteServer(t *testing.T) {
	// not found
	r1 := &mockRepo{getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) {
		return nil, gorm.ErrRecordNotFound
	}}
	uc1 := newUseCase(r1, &mockXLSX{})
	if err := uc1.DeleteServer(context.Background(), "x"); !errors.Is(err, domain.ErrServerNotFound) {
		t.Fatalf("want not found, got %v", err)
	}
	// other error
	r2 := &mockRepo{getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) {
		return nil, fmt.Errorf("boom")
	}}
	uc2 := newUseCase(r2, &mockXLSX{})
	if err := uc2.DeleteServer(context.Background(), "x"); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want internal, got %v", err)
	}
	// success delete
	r3 := &mockRepo{getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) {
		return &entity.Server{ServerID: "x"}, nil
	}, deleteFn: func(ctx context.Context, id string) error { return nil }}
	uc3 := newUseCase(r3, &mockXLSX{})
	if err := uc3.DeleteServer(context.Background(), "x"); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// delete error
	r4 := &mockRepo{getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) {
		return &entity.Server{ServerID: "x"}, nil
	}, deleteFn: func(ctx context.Context, id string) error { return fmt.Errorf("boom") }}
	uc4 := newUseCase(r4, &mockXLSX{})
	if err := uc4.DeleteServer(context.Background(), "x"); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want internal, got %v", err)
	}
}

func TestViewServer_AndExport(t *testing.T) {
	servers := []*entity.Server{{ServerID: "a", ServerName: "A", IPv4: "1.1.1.1", IntervalTime: 1}, {ServerID: "b", ServerName: "B", IPv4: "1.1.1.2", IntervalTime: 2}}
	r := &mockRepo{getServersFn: func(ctx context.Context, f dto.ServerFilterOptions, p dto.ServerPaginationOptions) ([]*entity.Server, int, error) {
		return servers, len(servers), nil
	}}
	uc := newUseCase(r, &mockXLSX{})

	// ViewServer
	res, total, err := uc.ViewServer(context.Background(), dto.ServerFilterOptions{}, dto.ServerPaginationOptions{})
	if err != nil || total != 2 || len(res) != 2 {
		t.Fatalf("unexpected view result: res=%d total=%d err=%v", len(res), total, err)
	}

	// Export - run in temp working directory
	cwd, _ := os.Getwd()
	t.Cleanup(func() { _ = os.Chdir(cwd) })
	tmp := t.TempDir()
	_ = os.Chdir(tmp)
	path, err := uc.ExportServer(context.Background(), dto.ServerFilterOptions{}, dto.ServerPaginationOptions{})
	if err != nil {
		t.Fatalf("export error: %v", err)
	}
	if _, statErr := os.Stat(path); statErr != nil {
		t.Fatalf("export file missing: %v", statErr)
	}
}

func TestViewServer_ErrorAndExport_ErrorFromView(t *testing.T) {
	// View error
	r1 := &mockRepo{getServersFn: func(ctx context.Context, f dto.ServerFilterOptions, p dto.ServerPaginationOptions) ([]*entity.Server, int, error) {
		return nil, 0, fmt.Errorf("boom")
	}}
	uc1 := newUseCase(r1, &mockXLSX{})
	if _, _, err := uc1.ViewServer(context.Background(), dto.ServerFilterOptions{}, dto.ServerPaginationOptions{}); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want internal, got %v", err)
	}

	// Export surfaces error from ViewServer
	if _, err := uc1.ExportServer(context.Background(), dto.ServerFilterOptions{}, dto.ServerPaginationOptions{}); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want internal, got %v", err)
	}
}

func TestImportServer(t *testing.T) {
	// invalid header
	x1 := &mockXLSX{getRowsFn: func(file string) ([][]string, error) { return [][]string{{"wrong"}}, nil }, validateFn: func(row []string) error { return domain.ErrInvalidFile }}
	uc1 := newUseCase(&mockRepo{}, x1)
	if _, err := uc1.ImportServer(context.Background(), "whatever.xlsx"); !errors.Is(err, domain.ErrInvalidFile) {
		t.Fatalf("want invalid file, got %v", err)
	}

	// parse errors and some success in batches
	rows := [][]string{{"server_id", "server_name", "ipv4", "location", "os", "interval_time"}}
	for i := 0; i < 10; i++ {
		if i%2 == 0 {
			rows = append(rows, []string{"id" + fmt.Sprint(i), "name", "1.1.1.1", "loc", "linux", "5"})
		} else {
			rows = append(rows, []string{"", "bad", "", "", "", ""})
		}
	}
	x2 := &mockXLSX{
		getRowsFn:  func(file string) ([][]string, error) { return rows, nil },
		validateFn: func(row []string) error { return nil },
		parseFn: func(row []string) (*entity.Server, error) {
			if row[0] == "" {
				return nil, fmt.Errorf("bad row")
			}
			return &entity.Server{ServerID: row[0], ServerName: row[1], IPv4: row[2], Location: row[3], OS: row[4], IntervalTime: 5}, nil
		},
	}
	r := &mockRepo{batchCreateFn: func(ctx context.Context, servers []*entity.Server) ([]*string, error) {
		ids := make([]*string, 0)
		for _, s := range servers {
			if s.ServerID != "id2" {
				ids = append(ids, &s.ServerID)
			}
		}
		return ids, nil
	}}
	uc2 := newUseCase(r, x2)
	resp, err := uc2.ImportServer(context.Background(), "file.xlsx")
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if resp.SuccessCount == 0 || resp.FailedCount == 0 {
		t.Fatalf("want some success and some failure, got %+v", resp)
	}

	// invalid file from GetRows
	x3 := &mockXLSX{getRowsFn: func(file string) ([][]string, error) { return nil, domain.ErrInvalidFile }}
	uc3 := newUseCase(&mockRepo{}, x3)
	if _, err := uc3.ImportServer(context.Background(), "bad.xlsx"); !errors.Is(err, domain.ErrInvalidFile) {
		t.Fatalf("want invalid file, got %v", err)
	}
}

func TestUpdateServer(t *testing.T) {
	nameTaken := "taken"
	newName := "new"
	base := &entity.Server{ServerID: "x", ServerName: "old", IPv4: "1.1.1.1", IntervalTime: 1}

	// not found
	r1 := &mockRepo{getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) {
		return nil, gorm.ErrRecordNotFound
	}}
	uc1 := newUseCase(r1, &mockXLSX{})
	if _, err := uc1.UpdateServer(context.Background(), "x", dto.UpdateServerParams{}); !errors.Is(err, domain.ErrServerNotFound) {
		t.Fatalf("want not found, got %v", err)
	}

	// get error
	r2 := &mockRepo{getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) {
		return nil, fmt.Errorf("boom")
	}}
	uc2 := newUseCase(r2, &mockXLSX{})
	if _, err := uc2.UpdateServer(context.Background(), "x", dto.UpdateServerParams{}); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want internal, got %v", err)
	}

	// name exists on other
	r3 := &mockRepo{
		getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) {
			if f == "server_id" {
				return &entity.Server{ServerID: "x"}, nil
			}
			// called to check name
			return &entity.Server{ServerID: "y"}, nil
		},
	}
	uc3 := newUseCase(r3, &mockXLSX{})
	if _, err := uc3.UpdateServer(context.Background(), "x", dto.UpdateServerParams{ServerName: &nameTaken}); !errors.Is(err, domain.ErrServerExist) {
		t.Fatalf("want exist, got %v", err)
	}

	// name check returns other error (not gorm.ErrRecordNotFound)
	r4 := &mockRepo{
		getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) {
			if f == "server_id" {
				return &entity.Server{ServerID: "x"}, nil
			}
			return nil, fmt.Errorf("boom")
		},
	}
	uc4 := newUseCase(r4, &mockXLSX{})
	if _, err := uc4.UpdateServer(context.Background(), "x", dto.UpdateServerParams{ServerName: &newName}); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want internal, got %v", err)
	}

	// success update
	updatedIPv4 := "2.2.2.2"
	r5 := &mockRepo{
		getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) { return base, nil },
		updateFn:     func(ctx context.Context, s *entity.Server) error { return nil },
	}
	uc5 := newUseCase(r5, &mockXLSX{})
	got, err := uc5.UpdateServer(context.Background(), "x", dto.UpdateServerParams{ServerName: &newName, IPv4: &updatedIPv4})
	if err != nil {
		t.Fatalf("unexpected err: %v", err)
	}
	if got.ServerName != newName || got.IPv4 != updatedIPv4 {
		t.Fatalf("not updated: %+v", got)
	}

	// update error path
	r6 := &mockRepo{
		getByFieldFn: func(ctx context.Context, f string, v interface{}) (*entity.Server, error) {
			return &entity.Server{ServerID: "x"}, nil
		},
		updateFn: func(ctx context.Context, s *entity.Server) error { return fmt.Errorf("boom") },
	}
	uc6 := newUseCase(r6, &mockXLSX{})
	if _, err := uc6.UpdateServer(context.Background(), "x", dto.UpdateServerParams{ServerName: &newName}); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want internal, got %v", err)
	}
}

func TestUpdateStatus(t *testing.T) {
	called := false
	r := &mockRepo{updateStatusFn: func(ctx context.Context, id string, st entity.ServerStatus) error { called = true; return nil }}
	uc := newUseCase(r, &mockXLSX{})
	if err := uc.UpdateStatus(context.Background(), dto.UpdateStatusMessage{ServerID: "x", Status: entity.ServerStatusOnline}); err != nil {
		t.Fatalf("unexpected: %v", err)
	}
	if !called {
		t.Fatalf("update not called")
	}

	r2 := &mockRepo{updateStatusFn: func(ctx context.Context, id string, st entity.ServerStatus) error { return fmt.Errorf("boom") }}
	uc2 := newUseCase(r2, &mockXLSX{})
	if err := uc2.UpdateStatus(context.Background(), dto.UpdateStatusMessage{ServerID: "x", Status: entity.ServerStatusOnline}); !errors.Is(err, domain.ErrInternalServer) {
		t.Fatalf("want internal, got %v", err)
	}
}

// Sanity to ensure excelize imported for ExportServer path coverage (avoid prune by compiler)
func TestExcelizeNewFile(t *testing.T) {
	f := excelize.NewFile()
	// write something to ensure it works in test env
	tmp := t.TempDir()
	p := filepath.Join(tmp, "tmp.xlsx")
	if err := f.SaveAs(p); err != nil {
		t.Fatalf("excelize save: %v", err)
	}
}
