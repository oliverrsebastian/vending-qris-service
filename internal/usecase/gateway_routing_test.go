package usecase_test

import (
	"context"
	"errors"
	"testing"

	"vending-qris-service/internal/usecase"
)

type mockPriorityRepo struct {
	list    []string
	listErr error
	replace []string
	rotated bool
}

func (m *mockPriorityRepo) ListOrdered(context.Context) ([]string, error) {
	if m.listErr != nil {
		return nil, m.listErr
	}
	return m.list, nil
}

func (m *mockPriorityRepo) ReplaceAll(_ context.Context, gateways []string) error {
	m.replace = gateways
	m.list = gateways
	return nil
}

func (m *mockPriorityRepo) RotateOnce(context.Context) error {
	if len(m.list) < 2 {
		return nil
	}
	m.rotated = true
	m.list = append(append([]string{}, m.list[1:]...), m.list[0])
	return nil
}

func TestGatewayRoutingUsecase_ListAndActive(t *testing.T) {
	uc := usecase.NewGatewayRoutingUsecase(
		&mockPriorityRepo{list: []string{"stub", "stub_fallback"}},
		&mockResolver{gw: stubGateway{name: "stub"}},
	)
	priority, active, err := uc.ListAndActive(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(priority) != 2 || active != "stub" {
		t.Fatalf("priority=%v active=%q", priority, active)
	}
}

func TestGatewayRoutingUsecase_SetPriority(t *testing.T) {
	repo := &mockPriorityRepo{list: []string{"stub"}}
	uc := usecase.NewGatewayRoutingUsecase(repo, &mockResolver{gw: stubGateway{name: "stub_fallback"}})
	if err := uc.SetPriority(context.Background(), []string{"stub_fallback", "stub"}); err != nil {
		t.Fatal(err)
	}
	if len(repo.replace) != 2 || repo.replace[0] != "stub_fallback" {
		t.Fatalf("replace: %v", repo.replace)
	}
}

func TestGatewayRoutingUsecase_FailoverRotate(t *testing.T) {
	repo := &mockPriorityRepo{list: []string{"a", "b", "c"}}
	uc := usecase.NewGatewayRoutingUsecase(repo, &mockResolver{gw: stubGateway{name: "b"}})
	priority, active, err := uc.FailoverRotate(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !repo.rotated || priority[0] != "b" || active != "b" {
		t.Fatalf("rotated=%v priority=%v active=%q", repo.rotated, priority, active)
	}
}

func TestGatewayRoutingUsecase_ListAndActive_resolverError(t *testing.T) {
	uc := usecase.NewGatewayRoutingUsecase(
		&mockPriorityRepo{list: []string{"stub"}},
		&mockResolver{err: errors.New("all down")},
	)
	_, _, err := uc.ListAndActive(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
}
