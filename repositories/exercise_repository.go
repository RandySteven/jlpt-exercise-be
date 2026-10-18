package repositories

import (
	"context"
	"encoding/json"

	mysql_client "github.com/RandySteven/go-cook/db"
	repository_interfaces "github.com/RandySteven/go-cook/db"
	"github.com/RandySteven/jlpt-exercise-be/entities/models"
	"github.com/RandySteven/jlpt-exercise-be/queries"
)

type (
	ExerciseRepository interface {
		repository_interfaces.Saver[models.Exercise]
		repository_interfaces.Finder[models.Exercise]
		repository_interfaces.Updater[models.Exercise]
		repository_interfaces.Deleter[models.Exercise]
		FindLatestByUserID(ctx context.Context, userID uint64) (result *models.Exercise, err error)
		UpdateStatusIfCurrent(ctx context.Context, id uint64, currentStatus, nextStatus, note string) (updated bool, err error)
	}

	exerciseRepository struct {
		dbx repository_interfaces.DBX
	}
)

func (e *exerciseRepository) Save(ctx context.Context, entity *models.Exercise) (result *models.Exercise, err error) {
	exerciseJSON := entity.Exercise
	if len(exerciseJSON) == 0 {
		exerciseJSON = json.RawMessage(`{}`)
	}
	payload, err := exerciseJSON.MarshalJSON()
	if err != nil {
		return nil, err
	}

	id, err := mysql_client.Save[models.Exercise](
		ctx,
		e.dbx(ctx),
		queries.InsertExercise.ToString(),
		entity.ReferenceID,
		entity.NLevel,
		string(payload),
		entity.Score,
		entity.Status,
		entity.UserID,
		entity.Note,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.ID = *id
	result.Exercise = exerciseJSON
	return result, nil
}

func (e *exerciseRepository) FindByID(ctx context.Context, id uint64) (result *models.Exercise, err error) {
	result = &models.Exercise{}
	err = e.dbx(ctx).QueryRowContext(ctx, queries.SelectExerciseByID.ToString(), id).Scan(
		&result.ID,
		&result.ReferenceID,
		&result.NLevel,
		&result.Exercise,
		&result.Score,
		&result.Status,
		&result.UserID,
		&result.Note,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (e *exerciseRepository) FindAll(ctx context.Context, skip uint64, take uint64) (result []*models.Exercise, err error) {
	rows, err := e.dbx(ctx).QueryContext(ctx, queries.SelectExercises.ToString(), take, skip)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		item := &models.Exercise{}
		if err = rows.Scan(
			&item.ID,
			&item.ReferenceID,
			&item.NLevel,
			&item.Exercise,
			&item.Score,
			&item.Status,
			&item.UserID,
			&item.Note,
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

func (e *exerciseRepository) Update(ctx context.Context, entity *models.Exercise) (result *models.Exercise, err error) {
	exerciseJSON := entity.Exercise
	if len(exerciseJSON) == 0 {
		exerciseJSON = json.RawMessage(`{}`)
	}
	payload, err := exerciseJSON.MarshalJSON()
	if err != nil {
		return nil, err
	}

	err = mysql_client.Update[models.Exercise](
		ctx,
		e.dbx(ctx),
		queries.UpdateExercise.ToString(),
		entity.ReferenceID,
		entity.NLevel,
		string(payload),
		entity.Score,
		entity.Status,
		entity.UserID,
		entity.Note,
		entity.ID,
	)
	if err != nil {
		return nil, err
	}
	result = entity
	result.Exercise = exerciseJSON
	return result, nil
}

func (e *exerciseRepository) DeleteByID(ctx context.Context, id uint64) (err error) {
	_, err = e.dbx(ctx).ExecContext(ctx, queries.DeleteExercise.ToString(), id)
	return err
}

func (e *exerciseRepository) FindLatestByUserID(ctx context.Context, userID uint64) (result *models.Exercise, err error) {
	result = &models.Exercise{}
	err = e.dbx(ctx).QueryRowContext(ctx, queries.SelectUserLatestExercise.ToString(), userID).Scan(
		&result.ID,
		&result.ReferenceID,
		&result.NLevel,
		&result.Exercise,
		&result.Score,
		&result.Status,
		&result.UserID,
		&result.Note,
		&result.CreatedAt,
		&result.UpdatedAt,
		&result.DeletedAt,
	)
	if err != nil {
		return nil, err
	}
	return result, nil
}

func (e *exerciseRepository) UpdateStatusIfCurrent(ctx context.Context, id uint64, currentStatus, nextStatus, note string) (updated bool, err error) {
	result, err := e.dbx(ctx).ExecContext(ctx, queries.UpdateExerciseStatusIfCurrent.ToString(), nextStatus, note, id, currentStatus)
	if err != nil {
		return false, err
	}
	affected, err := result.RowsAffected()
	if err != nil {
		return false, err
	}
	return affected > 0, nil
}

func newExerciseRepository(dbx repository_interfaces.DBX) *exerciseRepository {
	return &exerciseRepository{
		dbx: dbx,
	}
}

var _ ExerciseRepository = &exerciseRepository{}
