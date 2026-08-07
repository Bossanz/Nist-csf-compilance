package store

import (
	"context"

	"compliance/api/internal/domain"
	"github.com/jackc/pgx/v5"
)

const responseColumns = `
	r.id,r.project_id,r.subcategory_id,r.response_text,r.status,r.responded_by,
	r.submitted_at,r.review_comment,r.reviewed_by,r.reviewed_at,r.created_at,r.updated_at`
const responseReturningColumns = `
	id,project_id,subcategory_id,response_text,status,responded_by,
	submitted_at,review_comment,reviewed_by,reviewed_at,created_at,updated_at`

func (s *Store) ListResponses(ctx context.Context, projectID string) ([]StakeholderResponse, error) {
	if _, err := s.GetProject(ctx, projectID); err != nil {
		return nil, err
	}
	rows, err := s.DB.Query(ctx, `SELECT `+responseColumns+`
		FROM stakeholder_responses r
		JOIN subcategories sc ON sc.id=r.subcategory_id
		JOIN categories c ON c.id=sc.category_id
		JOIN functions f ON f.id=c.function_id
		WHERE r.project_id=$1
		ORDER BY f.code,c.code,sc.code`, projectID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	responses := []StakeholderResponse{}
	for rows.Next() {
		response, err := scanStakeholderResponse(rows)
		if err != nil {
			return nil, err
		}
		response.Documents, err = s.listResponseDocuments(ctx, response.ID)
		if err != nil {
			return nil, err
		}
		responses = append(responses, response)
	}
	return responses, rows.Err()
}

func (s *Store) CreateResponseDocument(ctx context.Context, projectID, subcategoryID, uploadedBy, originalName, storageKey, mimeType string, sizeBytes int64) (ResponseDocument, error) {
	var document ResponseDocument
	err := s.DB.QueryRow(ctx, `INSERT INTO response_documents(response_id,original_name,storage_key,mime_type,size_bytes,uploaded_by)
		SELECT r.id,$3,$4,$5,$6,$7
		FROM stakeholder_responses r
		WHERE r.project_id=$1 AND r.subcategory_id=$2
		RETURNING id,response_id,original_name,storage_key,mime_type,size_bytes,uploaded_by,created_at`,
		projectID, subcategoryID, originalName, storageKey, mimeType, sizeBytes, uploadedBy).Scan(responseDocumentArgs(&document)...)
	return document, err
}

func (s *Store) GetResponseDocument(ctx context.Context, projectID, subcategoryID, documentID string) (ResponseDocument, error) {
	var document ResponseDocument
	err := s.DB.QueryRow(ctx, `SELECT d.id,d.response_id,d.original_name,d.storage_key,d.mime_type,d.size_bytes,d.uploaded_by,d.created_at
		FROM response_documents d
		JOIN stakeholder_responses r ON r.id=d.response_id
		WHERE r.project_id=$1 AND r.subcategory_id=$2 AND d.id=$3`, projectID, subcategoryID, documentID).Scan(responseDocumentArgs(&document)...)
	return document, err
}

func (s *Store) DeleteResponseDocument(ctx context.Context, projectID, subcategoryID, documentID string) (ResponseDocument, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return ResponseDocument{}, err
	}
	defer tx.Rollback(ctx)

	var document ResponseDocument
	err = tx.QueryRow(ctx, `SELECT d.id,d.response_id,d.original_name,d.storage_key,d.mime_type,d.size_bytes,d.uploaded_by,d.created_at
		FROM response_documents d
		JOIN stakeholder_responses r ON r.id=d.response_id
		WHERE r.project_id=$1 AND r.subcategory_id=$2 AND d.id=$3
		FOR UPDATE`, projectID, subcategoryID, documentID).Scan(responseDocumentArgs(&document)...)
	if err != nil {
		return ResponseDocument{}, err
	}
	if _, err := tx.Exec(ctx, `DELETE FROM response_documents WHERE id=$1`, documentID); err != nil {
		return ResponseDocument{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ResponseDocument{}, err
	}
	return document, nil
}

func (s *Store) listResponseDocuments(ctx context.Context, responseID string) ([]ResponseDocument, error) {
	rows, err := s.DB.Query(ctx, `SELECT id,response_id,original_name,storage_key,mime_type,size_bytes,uploaded_by,created_at
		FROM response_documents WHERE response_id=$1 ORDER BY created_at,id`, responseID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	documents := []ResponseDocument{}
	for rows.Next() {
		var document ResponseDocument
		if err := rows.Scan(responseDocumentArgs(&document)...); err != nil {
			return nil, err
		}
		documents = append(documents, document)
	}
	return documents, rows.Err()
}

func (s *Store) SaveResponseDraft(ctx context.Context, projectID, subcategoryID, actorID, responseText string) (StakeholderResponse, error) {
	tx, err := s.DB.Begin(ctx)
	if err != nil {
		return StakeholderResponse{}, err
	}
	defer tx.Rollback(ctx)

	var currentStatus string
	err = tx.QueryRow(ctx, `SELECT status FROM stakeholder_responses WHERE project_id=$1 AND subcategory_id=$2 FOR UPDATE`, projectID, subcategoryID).Scan(&currentStatus)
	if err == pgx.ErrNoRows {
		var response StakeholderResponse
		err = tx.QueryRow(ctx, `INSERT INTO stakeholder_responses(project_id,subcategory_id,response_text,status,responded_by)
			VALUES ($1,$2,$3,$4,$5)
			RETURNING `+responseReturningColumns, projectID, subcategoryID, responseText, domain.ResponseDraft, actorID).Scan(responseArgs(&response)...)
		if err != nil {
			return StakeholderResponse{}, err
		}
		response.Documents = []ResponseDocument{}
		if err := tx.Commit(ctx); err != nil {
			return StakeholderResponse{}, err
		}
		return response, nil
	}
	if err != nil {
		return StakeholderResponse{}, err
	}
	if currentStatus != string(domain.ResponseDraft) && currentStatus != string(domain.ResponseNeedsMoreInfo) {
		return StakeholderResponse{}, domain.ErrInvalidResponseTransition
	}

	var response StakeholderResponse
	err = tx.QueryRow(ctx, `UPDATE stakeholder_responses
		SET response_text=$3, responded_by=$4, updated_at=now()
		WHERE project_id=$1 AND subcategory_id=$2
		RETURNING `+responseReturningColumns, projectID, subcategoryID, responseText, actorID).Scan(responseArgs(&response)...)
	if err != nil {
		return StakeholderResponse{}, err
	}
	response.Documents = []ResponseDocument{}
	if err := tx.Commit(ctx); err != nil {
		return StakeholderResponse{}, err
	}
	return response, nil
}

func (s *Store) SubmitResponse(ctx context.Context, projectID, subcategoryID, actorID string) (StakeholderResponse, error) {
	var response StakeholderResponse
	err := s.DB.QueryRow(ctx, `UPDATE stakeholder_responses
		SET status=$3, responded_by=$4, submitted_at=now(), reviewed_by=NULL, reviewed_at=NULL, updated_at=now()
		WHERE project_id=$1 AND subcategory_id=$2 AND status IN ($5,$6)
		RETURNING `+responseReturningColumns,
		projectID, subcategoryID, domain.ResponseSubmitted, actorID, domain.ResponseDraft, domain.ResponseNeedsMoreInfo).Scan(responseArgs(&response)...)
	if err == pgx.ErrNoRows {
		return StakeholderResponse{}, domain.ErrInvalidResponseTransition
	}
	if err != nil {
		return StakeholderResponse{}, err
	}
	response.Documents = []ResponseDocument{}
	return response, nil
}

func (s *Store) ReviewResponse(ctx context.Context, projectID, subcategoryID, actorID, status, comment string) (StakeholderResponse, error) {
	if !domain.CanTransitionResponse(domain.ResponseSubmitted, domain.ResponseStatus(status)) {
		return StakeholderResponse{}, domain.ErrInvalidResponseTransition
	}
	var response StakeholderResponse
	err := s.DB.QueryRow(ctx, `UPDATE stakeholder_responses
		SET status=$3, review_comment=$4, reviewed_by=$5, reviewed_at=now(), updated_at=now()
		WHERE project_id=$1 AND subcategory_id=$2 AND status=$6
		RETURNING `+responseReturningColumns,
		projectID, subcategoryID, status, comment, actorID, domain.ResponseSubmitted).Scan(responseArgs(&response)...)
	if err == pgx.ErrNoRows {
		return StakeholderResponse{}, domain.ErrInvalidResponseTransition
	}
	if err != nil {
		return StakeholderResponse{}, err
	}
	response.Documents = []ResponseDocument{}
	return response, nil
}

type responseRow interface {
	Scan(dest ...any) error
}

func scanStakeholderResponse(row responseRow) (StakeholderResponse, error) {
	var response StakeholderResponse
	err := row.Scan(responseArgs(&response)...)
	response.Documents = []ResponseDocument{}
	return response, err
}

func responseArgs(response *StakeholderResponse) []any {
	return []any{
		&response.ID,
		&response.ProjectID,
		&response.SubcategoryID,
		&response.ResponseText,
		&response.Status,
		&response.RespondedBy,
		&response.SubmittedAt,
		&response.ReviewComment,
		&response.ReviewedBy,
		&response.ReviewedAt,
		&response.CreatedAt,
		&response.UpdatedAt,
	}
}

func responseDocumentArgs(document *ResponseDocument) []any {
	return []any{
		&document.ID,
		&document.ResponseID,
		&document.OriginalName,
		&document.StorageKey,
		&document.MIMEType,
		&document.SizeBytes,
		&document.UploadedBy,
		&document.CreatedAt,
	}
}
