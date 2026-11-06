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
	ClassUserRepository interface {
		repository_interfaces.Saver[models.ClassUser]
		repository_interfaces.Finder[models.ClassUser]
		repository_interfaces.Updater[models.ClassUser]
		repository_interfaces.Deleter[models.ClassUser]
		FindByClassID(ctx context.Context, classID uint64) (result []*models.ClassUser, err error)
		FindByUserID(ctx context.Context, userID uint64) (result []*models.ClassUser, err error)
		FindByClassIDAndUserID(ctx context.Context, classID uint64, userID uint64) (result *models.ClassUser, err error)
	}

	classUserRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (c *classUserRepository) Save(ctx context.Context, entity *models.ClassUser) (result *models.ClassUser, err error) {
	id, err := mysql_client.Save[models.ClassUser](ctx, c.dbx(ctx), queries.InsertClassUser.ToString(), entity.ClassID, entity.UserID)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	return result, nil
}

func (c *classUserRepository) FindByID(ctx context.Context, id uint64) (result *models.ClassUser, err error) {
	return c.findOne(ctx, queries.SelectClassUserByID.ToString(), id)
}

func (c *classUserRepository) FindByClassIDAndUserID(ctx context.Context, classID uint64, userID uint64) (result *models.ClassUser, err error) {
	result = &models.ClassUser{}
	err = c.scan(c.dbx(ctx).QueryRowContext(ctx, queries.SelectClassUserByClassIDAndUserID.ToString(), classID, userID), result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *classUserRepository) FindByClassID(ctx context.Context, classID uint64) (result []*models.ClassUser, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClassUsersByClassID.ToString(), classID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClassUsers(rows)
}

func (c *classUserRepository) FindByUserID(ctx context.Context, userID uint64) (result []*models.ClassUser, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClassUsersByUserID.ToString(), userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClassUsers(rows)
}

func (c *classUserRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.ClassUser, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClassUsers.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClassUsers(rows)
}

func (c *classUserRepository) Update(ctx context.Context, entity *models.ClassUser) (result *models.ClassUser, err error) {
	err = mysql_client.Update[models.ClassUser](
		ctx,
		c.dbx(ctx),
		queries.UpdateClassUser.ToString(),
		entity.ClassID,
		entity.UserID,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (c *classUserRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = c.dbx(ctx).ExecContext(ctx, queries.DeleteClassUser.ToString(), id)
	return err
}

func (c *classUserRepository) findOne(ctx context.Context, query string, arg any) (result *models.ClassUser, err error) {
	result = &models.ClassUser{}
	err = c.scan(c.dbx(ctx).QueryRowContext(ctx, query, arg), result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *classUserRepository) scan(row interface{ Scan(...any) error }, result *models.ClassUser) error {
	return row.Scan(
		&result.ID,
		&result.ClassID,
		&result.UserID,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
}

func scanClassUsers(rows *sql.Rows) (result []*models.ClassUser, err error) {
	result = make([]*models.ClassUser, 0)
	for rows.Next() {
		item := &models.ClassUser{}
		if err = rows.Scan(
			&item.ID,
			&item.ClassID,
			&item.UserID,
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

func newClassUserRepository(dbx repository_interfaces.DBX) *classUserRepository {
	return &classUserRepository{
		dbx: dbx,
	}
}

var _ ClassUserRepository = &classUserRepository{}
