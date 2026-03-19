package rvinteraction

import (
	"testing"

	"github.com/Meesho/go-core/circuitbreaker"
	cohortConf "github.com/Meesho/iop-starter/cohort/config"
	"github.com/Meesho/rv-iop/internal/api/handler"
	"github.com/Meesho/rv-iop/internal/config"
	"github.com/stretchr/testify/assert"
)

type mockIopConfigHandler struct{}

func (m *mockIopConfigHandler) GetConfig(_ *cohortConf.IopConfigRequest) (*cohortConf.IopConfigResponse, error) {
	return nil, nil
}

func (m *mockIopConfigHandler) GetCB(_, _, _, _ string) (circuitbreaker.ManualCircuitBreaker, error) {
	return nil, nil
}

func TestNewStandardRvInteractionsWidgetImpl_ReturnsStructWhenAllDepsProvided(t *testing.T) {
	svc := &config.Service{}
	svc.App.Name = "test"
	iopHandler := &mockIopConfigHandler{}

	impl := NewStandardRvInteractionsWidgetImpl(svc, iopHandler)

	assert.NotNil(t, impl)
	assert.Equal(t, svc, impl.ServiceConf)
	assert.Equal(t, iopHandler, impl.IopConfigHandler)
}

func TestNewStandardRvInteractionsWidgetImpl_PanicsWhenServiceConfNil(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Contains(t, r, "service conf cannot be nil")
	}()
	NewStandardRvInteractionsWidgetImpl(nil, &mockIopConfigHandler{})
	t.Error("expected panic when serviceConf is nil")
}

func TestNewStandardRvInteractionsWidgetImpl_PanicsWhenIopConfigHandlerNil(t *testing.T) {
	defer func() {
		r := recover()
		assert.NotNil(t, r)
		assert.Contains(t, r, "iop config handler cannot be nil")
	}()
	NewStandardRvInteractionsWidgetImpl(&config.Service{}, nil)
	t.Error("expected panic when iopConfigHandler is nil")
}

func TestBuildCategoriesFromProducts_GroupsBySscatId(t *testing.T) {
	categoryLimit := 10
	productPerCategoryLimit := 5
	request := &GetRvInteractionsWidgetRequest{}
	request.Data.CategoryLimit = &categoryLimit
	request.Data.ProductPerCategoryLimit = &productPerCategoryLimit

	products := []Product{
		{
			RvInteractionProduct: handler.RvInteractionProduct{
				ProductId:  101,
				CatalogId:  1,
				MetaData:   &handler.MetaData{Context: map[string]string{"sscat_id": "10", "sscat_name": "Electronics"}},
			},
		},
		{
			RvInteractionProduct: handler.RvInteractionProduct{
				ProductId:  102,
				CatalogId:  2,
				MetaData:   &handler.MetaData{Context: map[string]string{"sscat_id": "10", "sscat_name": "Electronics"}},
			},
		},
		{
			RvInteractionProduct: handler.RvInteractionProduct{
				ProductId:  201,
				CatalogId:  3,
				MetaData:   &handler.MetaData{Context: map[string]string{"sscat_id": "20", "sscat_name": "Fashion"}},
			},
		},
	}

	impl := &StandardRvInteractionsWidgetImpl{}
	categories := impl.buildCategoriesFromProducts(products, request)

	assert.Len(t, categories, 2)
	assert.Equal(t, 10, categories[0].SscatId)
	assert.Equal(t, "Electronics", categories[0].SscatName)
	assert.Equal(t, []int{101, 102}, categories[0].ProductIds)
	assert.Equal(t, 20, categories[1].SscatId)
	assert.Equal(t, "Fashion", categories[1].SscatName)
	assert.Equal(t, []int{201}, categories[1].ProductIds)
}

func TestBuildCategoriesFromProducts_SkipsProductsWithZeroSscatId(t *testing.T) {
	request := &GetRvInteractionsWidgetRequest{}

	products := []Product{
		{
			RvInteractionProduct: handler.RvInteractionProduct{
				ProductId: 101,
				CatalogId: 1,
				MetaData:  &handler.MetaData{Context: map[string]string{"sscat_id": "0", "sscat_name": "Unknown"}},
			},
		},
		{
			RvInteractionProduct: handler.RvInteractionProduct{
				ProductId: 102,
				CatalogId: 2,
				MetaData:  nil,
			},
		},
		{
			RvInteractionProduct: handler.RvInteractionProduct{
				ProductId: 103,
				CatalogId: 3,
				MetaData:  &handler.MetaData{Context: map[string]string{"sscat_id": "10", "sscat_name": "Electronics"}},
			},
		},
	}

	impl := &StandardRvInteractionsWidgetImpl{}
	categories := impl.buildCategoriesFromProducts(products, request)

	assert.Len(t, categories, 1)
	assert.Equal(t, 10, categories[0].SscatId)
	assert.Equal(t, []int{103}, categories[0].ProductIds)
}

func TestBuildCategoriesFromProducts_AppliesProductPerCategoryLimit(t *testing.T) {
	productPerCategoryLimit := 2
	request := &GetRvInteractionsWidgetRequest{}
	request.Data.ProductPerCategoryLimit = &productPerCategoryLimit

	products := []Product{
		{RvInteractionProduct: handler.RvInteractionProduct{ProductId: 101, CatalogId: 1, MetaData: &handler.MetaData{Context: map[string]string{"sscat_id": "10", "sscat_name": "A"}}}},
		{RvInteractionProduct: handler.RvInteractionProduct{ProductId: 102, CatalogId: 2, MetaData: &handler.MetaData{Context: map[string]string{"sscat_id": "10", "sscat_name": "A"}}}},
		{RvInteractionProduct: handler.RvInteractionProduct{ProductId: 103, CatalogId: 3, MetaData: &handler.MetaData{Context: map[string]string{"sscat_id": "10", "sscat_name": "A"}}}},
		{RvInteractionProduct: handler.RvInteractionProduct{ProductId: 104, CatalogId: 4, MetaData: &handler.MetaData{Context: map[string]string{"sscat_id": "10", "sscat_name": "A"}}}},
	}

	impl := &StandardRvInteractionsWidgetImpl{}
	categories := impl.buildCategoriesFromProducts(products, request)

	assert.Len(t, categories, 1)
	assert.Equal(t, []int{101, 102}, categories[0].ProductIds)
}

func TestBuildCategoriesFromProducts_AppliesCategoryLimit(t *testing.T) {
	categoryLimit := 2
	request := &GetRvInteractionsWidgetRequest{}
	request.Data.CategoryLimit = &categoryLimit

	products := []Product{
		{RvInteractionProduct: handler.RvInteractionProduct{ProductId: 1, CatalogId: 1, MetaData: &handler.MetaData{Context: map[string]string{"sscat_id": "10", "sscat_name": "A"}}}},
		{RvInteractionProduct: handler.RvInteractionProduct{ProductId: 2, CatalogId: 2, MetaData: &handler.MetaData{Context: map[string]string{"sscat_id": "20", "sscat_name": "B"}}}},
		{RvInteractionProduct: handler.RvInteractionProduct{ProductId: 3, CatalogId: 3, MetaData: &handler.MetaData{Context: map[string]string{"sscat_id": "30", "sscat_name": "C"}}}},
	}

	impl := &StandardRvInteractionsWidgetImpl{}
	categories := impl.buildCategoriesFromProducts(products, request)

	assert.Len(t, categories, 2)
	assert.Equal(t, 10, categories[0].SscatId)
	assert.Equal(t, 20, categories[1].SscatId)
}

func TestBuildCategoriesFromProducts_ReturnsEmptyWhenNoProducts(t *testing.T) {
	impl := &StandardRvInteractionsWidgetImpl{}
	categories := impl.buildCategoriesFromProducts(nil, &GetRvInteractionsWidgetRequest{})
	assert.Empty(t, categories)

	categories = impl.buildCategoriesFromProducts([]Product{}, &GetRvInteractionsWidgetRequest{})
	assert.Empty(t, categories)
}
