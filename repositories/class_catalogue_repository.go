package repositories

import (
	"context"
	"database/sql"

	mysql_client "github.com/RandySteven/go-cook/db"
	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	ClassCatalogueRepository interface {
		repository_interfaces.Saver[models.ClassCatalogue]
		repository_interfaces.Finder[models.ClassCatalogue]
		repository_interfaces.Updater[models.ClassCatalogue]
		repository_interfaces.Deleter[models.ClassCatalogue]
		FindByCategoryID(ctx context.Context, categoryID uint64) (result []*models.ClassCatalogue, err error)
	}

	classCatalogueRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (c *classCatalogueRepository) Save(ctx context.Context, entity *models.ClassCatalogue) (result *models.ClassCatalogue, err error) {
	additionalInfo, additionalInfoPayload, err := jsonColumn(entity.AdditionalInfo)
	if err != nil {
		return nil, err
	}

	id, err := mysql_client.Save[models.ClassCatalogue](
		ctx,
		c.dbx(ctx),
		queries.InsertClassCatalogue.ToString(),
		entity.Name,
		entity.CategoryID,
		entity.Description,
		additionalInfoPayload,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	result.AdditionalInfo = additionalInfo
	return result, nil
}

func (c *classCatalogueRepository) FindByID(ctx context.Context, id uint64) (result *models.ClassCatalogue, err error) {
	return c.findOne(ctx, queries.SelectClassCatalogueByID.ToString(), id)
}

func (c *classCatalogueRepository) FindByCategoryID(ctx context.Context, categoryID uint64) (result []*models.ClassCatalogue, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClassCataloguesByCategoryID.ToString(), categoryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClassCatalogues(rows)
}

func (c *classCatalogueRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.ClassCatalogue, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClassCatalogues.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClassCatalogues(rows)
}

func (c *classCatalogueRepository) Update(ctx context.Context, entity *models.ClassCatalogue) (result *models.ClassCatalogue, err error) {
	additionalInfo, additionalInfoPayload, err := jsonColumn(entity.AdditionalInfo)
	if err != nil {
		return nil, err
	}

	err = mysql_client.Update[models.ClassCatalogue](
		ctx,
		c.dbx(ctx),
		queries.UpdateClassCatalogue.ToString(),
		entity.Name,
		entity.CategoryID,
		entity.Description,
		additionalInfoPayload,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.AdditionalInfo = additionalInfo
	return result, nil
}

func (c *classCatalogueRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = c.dbx(ctx).ExecContext(ctx, queries.DeleteClassCatalogue.ToString(), id)
	return err
}

func (c *classCatalogueRepository) findOne(ctx context.Context, query string, arg any) (result *models.ClassCatalogue, err error) {
	result = &models.ClassCatalogue{}
	err = c.dbx(ctx).QueryRowContext(ctx, query, arg).Scan(
		&result.ID,
		&result.Name,
		&result.CategoryID,
		&result.Description,
		&result.AdditionalInfo,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func scanClassCatalogues(rows *sql.Rows) (result []*models.ClassCatalogue, err error) {
	result = make([]*models.ClassCatalogue, 0)
	for rows.Next() {
		item := &models.ClassCatalogue{}
		if err = rows.Scan(
			&item.ID,
			&item.Name,
			&item.CategoryID,
			&item.Description,
			&item.AdditionalInfo,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func newClassCatalogueRepository(dbx repository_interfaces.DBX) *classCatalogueRepository {
	return &classCatalogueRepository{
		dbx: dbx,
	}
}

var _ ClassCatalogueRepository = &classCatalogueRepository{}
