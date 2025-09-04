package services

import (
	"context"
	"database/sql"
	"fmt"
	"time"
	"url-manager-system/backend/internal/db/models"

	"github.com/google/uuid"
	"github.com/sirupsen/logrus"
)

// ProjectService 项目服务
type ProjectService struct {
	db *sql.DB
}

// NewProjectService 创建项目服务
func NewProjectService(db *sql.DB) *ProjectService {
	return &ProjectService{db: db}
}

// CreateProject 创建项目
func (s *ProjectService) CreateProject(ctx context.Context, name, description string) (*models.Project, error) {
	project := &models.Project{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}

	query := `
		INSERT INTO projects (id, name, description, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id, name, description, created_at, updated_at
	`

	err := s.db.QueryRowContext(ctx, query,
		project.ID, project.Name, project.Description, project.CreatedAt, project.UpdatedAt,
	).Scan(&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt)

	if err != nil {
		logrus.WithError(err).Error("Failed to create project")
		return nil, fmt.Errorf("failed to create project: %w", err)
	}

	logrus.WithField("project_id", project.ID).Info("Project created successfully")
	return project, nil
}

// GetProject 获取项目
func (s *ProjectService) GetProject(ctx context.Context, id uuid.UUID) (*models.Project, error) {
	project := &models.Project{}
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM projects
		WHERE id = $1
	`

	err := s.db.QueryRowContext(ctx, query, id).Scan(
		&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		logrus.WithError(err).Error("Failed to get project")
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// GetProjectByName 根据名称获取项目
func (s *ProjectService) GetProjectByName(ctx context.Context, name string) (*models.Project, error) {
	project := &models.Project{}
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM projects
		WHERE name = $1
	`

	err := s.db.QueryRowContext(ctx, query, name).Scan(
		&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		logrus.WithError(err).Error("Failed to get project by name")
		return nil, fmt.Errorf("failed to get project: %w", err)
	}

	return project, nil
}

// ListProjects 列出项目
func (s *ProjectService) ListProjects(ctx context.Context, limit, offset int) ([]models.Project, int, error) {
	// 获取总数
	var total int
	err := s.db.QueryRowContext(ctx, "SELECT COUNT(*) FROM projects").Scan(&total)
	if err != nil {
		logrus.WithError(err).Error("Failed to count projects")
		return nil, 0, fmt.Errorf("failed to count projects: %w", err)
	}

	// 获取项目列表
	query := `
		SELECT id, name, description, created_at, updated_at
		FROM projects
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := s.db.QueryContext(ctx, query, limit, offset)
	if err != nil {
		logrus.WithError(err).Error("Failed to list projects")
		return nil, 0, fmt.Errorf("failed to list projects: %w", err)
	}
	defer rows.Close()

	var projects []models.Project
	for rows.Next() {
		var project models.Project
		err := rows.Scan(
			&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt,
		)
		if err != nil {
			logrus.WithError(err).Error("Failed to scan project")
			continue
		}
		projects = append(projects, project)
	}

	if err = rows.Err(); err != nil {
		logrus.WithError(err).Error("Error iterating projects")
		return nil, 0, fmt.Errorf("error iterating projects: %w", err)
	}

	return projects, total, nil
}

// UpdateProject 更新项目
func (s *ProjectService) UpdateProject(ctx context.Context, id uuid.UUID, name, description string) (*models.Project, error) {
	query := `
		UPDATE projects 
		SET name = $2, description = $3, updated_at = $4
		WHERE id = $1
		RETURNING id, name, description, created_at, updated_at
	`

	project := &models.Project{}
	err := s.db.QueryRowContext(ctx, query, id, name, description, time.Now()).Scan(
		&project.ID, &project.Name, &project.Description, &project.CreatedAt, &project.UpdatedAt,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("project not found")
		}
		logrus.WithError(err).Error("Failed to update project")
		return nil, fmt.Errorf("failed to update project: %w", err)
	}

	logrus.WithField("project_id", project.ID).Info("Project updated successfully")
	return project, nil
}

// DeleteProject 删除项目
func (s *ProjectService) DeleteProject(ctx context.Context, id uuid.UUID) error {
	// 检查项目下是否还有活跃的URL
	var activeURLCount int
	countQuery := `
		SELECT COUNT(*) FROM ephemeral_urls 
		WHERE project_id = $1 AND status IN ('creating', 'active')
	`
	err := s.db.QueryRowContext(ctx, countQuery, id).Scan(&activeURLCount)
	if err != nil {
		logrus.WithError(err).Error("Failed to count active URLs")
		return fmt.Errorf("failed to count active URLs: %w", err)
	}

	if activeURLCount > 0 {
		return fmt.Errorf("cannot delete project with active URLs")
	}

	// 删除项目
	deleteQuery := `DELETE FROM projects WHERE id = $1`
	result, err := s.db.ExecContext(ctx, deleteQuery, id)
	if err != nil {
		logrus.WithError(err).Error("Failed to delete project")
		return fmt.Errorf("failed to delete project: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("project not found")
	}

	logrus.WithField("project_id", id).Info("Project deleted successfully")
	return nil
}
