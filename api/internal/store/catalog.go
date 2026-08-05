package store

import "context"

func (s *Store) ListFunctions(ctx context.Context) ([]Function, error) {
	rows, err := s.DB.Query(ctx, `SELECT f.id,f.code,f.name,f.description,c.id,c.code,c.name,c.description,sc.id,sc.code,sc.description
FROM functions f LEFT JOIN categories c ON c.function_id=f.id LEFT JOIN subcategories sc ON sc.category_id=c.id ORDER BY f.code,c.code,sc.code`)
	if err != nil { return nil, err }; defer rows.Close()
	byCode := map[string]int{}; out := []Function{}
	for rows.Next() {
		var f Function
		var cID, cCode, cName, cDesc, scID, scCode, scDesc *string
		if err := rows.Scan(&f.ID,&f.Code,&f.Name,&f.Description,&cID,&cCode,&cName,&cDesc,&scID,&scCode,&scDesc); err != nil { return nil, err }
		idx, ok := byCode[f.Code]; if !ok { out = append(out, f); idx = len(out)-1; byCode[f.Code] = idx }
		if cID == nil { continue }
		catIdx := -1; for i := range out[idx].Categories { if out[idx].Categories[i].Code == *cCode { catIdx = i; break } }
		if catIdx < 0 { out[idx].Categories = append(out[idx].Categories, Category{ID:*cID, FunctionID:f.ID, Code:*cCode, Name:*cName, Description:*cDesc}); catIdx = len(out[idx].Categories)-1 }
		if scID != nil { out[idx].Categories[catIdx].Subcategories = append(out[idx].Categories[catIdx].Subcategories, Subcategory{ID:*scID, CategoryID:*cID, Code:*scCode, Description:*scDesc}) }
	}
	return out, rows.Err()
}
