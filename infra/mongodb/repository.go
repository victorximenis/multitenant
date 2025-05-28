package mongodb

import (
	"context"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	"github.com/victorximenis/multitenant/core"
)

const (
	DATABASE_NAME   = "multitenant"
	COLLECTION_NAME = "tenants"
)

type TenantRepository struct {
	client     *mongo.Client
	collection *mongo.Collection
}

func NewTenantRepository(ctx context.Context, uri string) (*TenantRepository, error) {
	clientOptions := options.Client().ApplyURI(uri)
	client, err := mongo.Connect(ctx, clientOptions)
	if err != nil {
		return nil, err
	}

	// Verify connection
	if err := client.Ping(ctx, nil); err != nil {
		return nil, err
	}

	collection := client.Database(DATABASE_NAME).Collection(COLLECTION_NAME)

	// Create indexes
	indexModels := []mongo.IndexModel{
		{
			Keys:    bson.D{{Key: "name", Value: 1}},
			Options: options.Index().SetUnique(true),
		},
		{
			Keys: bson.D{{Key: "datasources.id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "id", Value: 1}},
		},
		{
			Keys: bson.D{{Key: "is_active", Value: 1}},
		},
	}

	_, err = collection.Indexes().CreateMany(ctx, indexModels)
	if err != nil {
		return nil, err
	}

	return &TenantRepository{
		client:     client,
		collection: collection,
	}, nil
}

func (r *TenantRepository) GetByName(ctx context.Context, name string) (*core.Tenant, error) {
	var tenant core.Tenant

	filter := bson.M{"name": name}
	err := r.collection.FindOne(ctx, filter).Decode(&tenant)

	if err != nil {
		if err == mongo.ErrNoDocuments {
			return nil, core.TenantNotFoundError{Name: name}
		}
		return nil, err
	}

	if !tenant.IsActive {
		return nil, core.TenantInactiveError{Name: name}
	}

	return &tenant, nil
}

func (r *TenantRepository) List(ctx context.Context) ([]core.Tenant, error) {
	var tenants []core.Tenant

	cursor, err := r.collection.Find(ctx, bson.M{})
	if err != nil {
		return nil, err
	}
	defer cursor.Close(ctx)

	if err := cursor.All(ctx, &tenants); err != nil {
		return nil, err
	}

	return tenants, nil
}

func (r *TenantRepository) Create(ctx context.Context, tenant *core.Tenant) error {
	tenant.CreatedAt = time.Now()
	tenant.UpdatedAt = time.Now()

	_, err := r.collection.InsertOne(ctx, tenant)
	return err
}

func (r *TenantRepository) Update(ctx context.Context, tenant *core.Tenant) error {
	tenant.UpdatedAt = time.Now()

	filter := bson.M{"id": tenant.ID}
	update := bson.M{"$set": tenant}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return core.TenantNotFoundError{Name: tenant.Name}
	}

	return nil
}

func (r *TenantRepository) Delete(ctx context.Context, id string) error {
	filter := bson.M{"id": id}

	result, err := r.collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}

	if result.DeletedCount == 0 {
		return core.TenantNotFoundError{Name: ""}
	}

	return nil
}

// AddDatasource adds a new datasource to an existing tenant
func (r *TenantRepository) AddDatasource(ctx context.Context, tenantID string, datasource core.Datasource) error {
	filter := bson.M{"id": tenantID}
	update := bson.M{
		"$push": bson.M{"datasources": datasource},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return core.TenantNotFoundError{Name: ""}
	}

	return nil
}

// RemoveDatasource removes a datasource from an existing tenant
func (r *TenantRepository) RemoveDatasource(ctx context.Context, tenantID, datasourceID string) error {
	filter := bson.M{"id": tenantID}
	update := bson.M{
		"$pull": bson.M{"datasources": bson.M{"id": datasourceID}},
		"$set":  bson.M{"updated_at": time.Now()},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return core.TenantNotFoundError{Name: ""}
	}

	return nil
}

// UpdateDatasource updates a specific datasource within a tenant
func (r *TenantRepository) UpdateDatasource(ctx context.Context, tenantID string, datasource core.Datasource) error {
	filter := bson.M{
		"id":             tenantID,
		"datasources.id": datasource.ID,
	}
	update := bson.M{
		"$set": bson.M{
			"datasources.$": datasource,
			"updated_at":    time.Now(),
		},
	}

	result, err := r.collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	if result.MatchedCount == 0 {
		return core.TenantNotFoundError{Name: ""}
	}

	return nil
}
