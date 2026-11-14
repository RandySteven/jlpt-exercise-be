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
	ClassCategoryRepository interface {
		repository_interfaces.Saver[models.ClassCategory]
		repository_interfaces.Finder[models.ClassCategory]
		repository_interfaces.Updater[models.ClassCategory]
		repository_interfaces.Deleter[models.ClassCategory]
		FindByCategory(ctx context.Context, category string) (result *models.ClassCategory, err error)
	}

	classCategoryRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (c *classCategoryRepository) Save(ctx context.Context, entity *models.ClassCategory) (result *models.ClassCategory, err error) {
	id, err := mysql_client.Save[models.ClassCategory](
		ctx,
		c.dbx(ctx),
		queries.InsertClassCategory.ToString(),
		entity.Category,
		entity.Description,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	return result, nil
}

func (c *classCategoryRepository) FindByID(ctx context.Context, id uint64) (result *models.ClassCategory, err error) {
	return c.findOne(ctx, queries.SelectClassCategoryByID.ToString(), id)
}

func (c *classCategoryRepository) FindByCategory(ctx context.Context, category string) (result *models.ClassCategory, err error) {
	return c.findOne(ctx, queries.SelectClassCategoryByCategory.ToString(), category)
}

func (c *classCategoryRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.ClassCategory, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClassCategories.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClassCategories(rows)
}

func (c *classCategoryRepository) Update(ctx context.Context, entity *models.ClassCategory) (result *models.ClassCategory, err error) {
	err = mysql_client.Update[models.ClassCategory](
		ctx,
		c.dbx(ctx),
		queries.UpdateClassCategory.ToString(),
		entity.Category,
		entity.Description,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (c *classCategoryRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = c.dbx(ctx).ExecContext(ctx, queries.DeleteClassCategory.ToString(), id)
	return err
}

func (c *classCategoryRepository) findOne(ctx context.Context, query string, arg any) (result *models.ClassCategory, err error) {
	result = &models.ClassCategory{}
	err = c.dbx(ctx).QueryRowContext(ctx, query, arg).Scan(
		&result.ID,
		&result.Category,
		&result.Description,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func scanClassCategories(rows *sql.Rows) (result []*models.ClassCategory, err error) {
	result = make([]*models.ClassCategory, 0)
	for rows.Next() {
		item := &models.ClassCategory{}
		if err = rows.Scan(
			&item.ID,
			&item.Category,
			&item.Description,
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

func newClassCategoryRepository(dbx repository_interfaces.DBX) *classCategoryRepository {
	return &classCategoryRepository{
		dbx: dbx,
	}
}

var _ ClassCategoryRepository = &classCategoryRepository{}
