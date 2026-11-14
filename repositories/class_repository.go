package repositories

import (
	"context"
	"database/sql"
	"encoding/json"

	mysql_client "github.com/RandySteven/go-cook/db"
	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	ClassRepository interface {
		repository_interfaces.Saver[models.Class]
		repository_interfaces.Finder[models.Class]
		repository_interfaces.Updater[models.Class]
		repository_interfaces.Deleter[models.Class]
		FindBySenseiID(ctx context.Context, senseiID uint64) (result []*models.Class, err error)
		FindByReferenceID(ctx context.Context, referenceID string) (result *models.Class, err error)
	}

	classRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (c *classRepository) Save(ctx context.Context, entity *models.Class) (result *models.Class, err error) {
	schedules, err := schedulesPayload(entity.Schedules)
	if err != nil {
		return nil, err
	}

	id, err := mysql_client.Save[models.Class](ctx, c.dbx(ctx), queries.InsertClass.ToString(), entity.SenseiID, schedules)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	result.Schedules, err = decodeSchedules([]byte(schedules))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *classRepository) FindByID(ctx context.Context, id uint64) (result *models.Class, err error) {
	return c.findOne(ctx, queries.SelectClassByID.ToString(), id)
}

func (c *classRepository) FindByReferenceID(ctx context.Context, referenceID string) (result *models.Class, err error) {
	return c.findOne(ctx, ``, referenceID)
}

func (c *classRepository) FindBySenseiID(ctx context.Context, senseiID uint64) (result []*models.Class, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClassesBySenseiID.ToString(), senseiID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClasses(rows)
}

func (c *classRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.Class, err error) {
	rows, err := c.dbx(ctx).QueryContext(ctx, queries.SelectClasses.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanClasses(rows)
}

func (c *classRepository) Update(ctx context.Context, entity *models.Class) (result *models.Class, err error) {
	schedules, err := schedulesPayload(entity.Schedules)
	if err != nil {
		return nil, err
	}

	err = mysql_client.Update[models.Class](
		ctx,
		c.dbx(ctx),
		queries.UpdateClass.ToString(),
		entity.SenseiID,
		schedules,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.Schedules, err = decodeSchedules([]byte(schedules))
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (c *classRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = c.dbx(ctx).ExecContext(ctx, queries.DeleteClass.ToString(), id)
	return err
}

func (c *classRepository) findOne(ctx context.Context, query string, arg any) (result *models.Class, err error) {
	result = &models.Class{}
	var schedules []byte
	err = c.dbx(ctx).QueryRowContext(ctx, query, arg).Scan(
		&result.ID,
		&result.SenseiID,
		&schedules,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	result.Schedules, err = decodeSchedules(schedules)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func scanClasses(rows *sql.Rows) (result []*models.Class, err error) {
	result = make([]*models.Class, 0)
	for rows.Next() {
		item := &models.Class{}
		var schedules []byte
		if err = rows.Scan(
			&item.ID,
			&item.SenseiID,
			&schedules,
			&item.CreatedAt,
			&item.UpdatedAt,
			&item.DeletedAt,
		); err != nil {
			return nil, err
		}
		item.Schedules, err = decodeSchedules(schedules)
		if err != nil {
			return nil, err
		}
		result = append(result, item)
	}
	if err = rows.Err(); err != nil {
		return nil, err
	}
	return result, nil
}

func schedulesPayload(schedules []models.ClassSchedule) (string, error) {
	if schedules == nil {
		schedules = []models.ClassSchedule{}
	}
	payload, err := json.Marshal(schedules)
	if err != nil {
		return "", err
	}
	return string(payload), nil
}

func decodeSchedules(raw []byte) ([]models.ClassSchedule, error) {
	if len(raw) == 0 {
		return []models.ClassSchedule{}, nil
	}
	schedules := make([]models.ClassSchedule, 0)
	if err := json.Unmarshal(raw, &schedules); err != nil {
		return nil, err
	}
	return schedules, nil
}

func newClassRepository(dbx repository_interfaces.DBX) *classRepository {
	return &classRepository{
		dbx: dbx,
	}
}

var _ ClassRepository = &classRepository{}
