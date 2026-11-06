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
	ClassRoomRepository interface {
		repository_interfaces.Saver[models.ClassRoom]
		repository_interfaces.Finder[models.ClassRoom]
		repository_interfaces.Updater[models.ClassRoom]
		repository_interfaces.Deleter[models.ClassRoom]
		FindByRoomCode(ctx context.Context, roomCode string) (result *models.ClassRoom, err error)
		FindByCreatedBy(ctx context.Context, createdBy uint64) (result []*models.ClassRoom, err error)
	}

	classRoomRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (c *classRoomRepository) Save(ctx context.Context, entity *models.ClassRoom) (result *models.ClassRoom, err error) {
	id, err := mysql_client.Save[models.ClassRoom](
		ctx,
		c.dbx(ctx),
		queries.InsertClassRoom.ToString(),
		entity.ClassRoomName,
		entity.RoomCode,
		entity.CreatedBy,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	return result, nil
}

func (c *classRoomRepository) FindByID(ctx context.Context, id uint64) (result *models.ClassRoom, err error) {
	return c.findOne(ctx, queries.SelectClassRoomByID.ToString(), id)
}

func (c *classRoomRepository) FindByRoomCode(ctx context.Context, roomCode string) (result *models.ClassRoom, err error) {
	return c.findOne(ctx, queries.SelectClassRoomByRoomCode.ToString(), roomCode)
}

func (c *classRoomRepository) FindByCreatedBy(ctx context.Context, createdBy uint64) (result []*models.ClassRoom, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClassRoomsByCreatedBy.ToString(), createdBy)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClassRooms(rows)
}

func (c *classRoomRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.ClassRoom, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClassRooms.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClassRooms(rows)
}

func (c *classRoomRepository) Update(ctx context.Context, entity *models.ClassRoom) (result *models.ClassRoom, err error) {
	err = mysql_client.Update[models.ClassRoom](
		ctx,
		c.dbx(ctx),
		queries.UpdateClassRoom.ToString(),
		entity.ClassRoomName,
		entity.RoomCode,
		entity.CreatedBy,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	return entity, nil
}

func (c *classRoomRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = c.dbx(ctx).ExecContext(ctx, queries.DeleteClassRoom.ToString(), id)
	return err
}

func (c *classRoomRepository) findOne(ctx context.Context, query string, arg any) (result *models.ClassRoom, err error) {
	result = &models.ClassRoom{}
	err = c.dbx(ctx).QueryRowContext(ctx, query, arg).Scan(
		&result.ID,
		&result.ClassRoomName,
		&result.RoomCode,
		&result.CreatedBy,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func scanClassRooms(rows *sql.Rows) (result []*models.ClassRoom, err error) {
	result = make([]*models.ClassRoom, 0)
	for rows.Next() {
		item := &models.ClassRoom{}
		if err = rows.Scan(
			&item.ID,
			&item.ClassRoomName,
			&item.RoomCode,
			&item.CreatedBy,
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

func newClassRoomRepository(dbx repository_interfaces.DBX) *classRoomRepository {
	return &classRoomRepository{
		dbx: dbx,
	}
}

var _ ClassRoomRepository = &classRoomRepository{}
