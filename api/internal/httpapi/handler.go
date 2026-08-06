package httpapi

import (
	"context"
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"strings"
	"compliance/api/internal/domain"
	"compliance/api/internal/store"
)

type dataStore interface {
	ListFunctions(context.Context) ([]store.Function,error); CreateProject(context.Context,string,string) (store.Project,error); ListProjects(context.Context) ([]store.Project,error); GetProject(context.Context,string) (store.Project,error); ListProfile(context.Context,string) ([]store.ProfileRow,error); UpdateProfile(context.Context,string,string,store.ProfilePatch) (store.ProfileRow,error)
}
type Handler struct { Store dataStore }
type errorBody struct { Error struct { Code string `json:"code"`; Message string `json:"message"` } `json:"error"` }

func New(s *store.Store) http.Handler { return &Handler{Store:s} }
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*"); w.Header().Set("Access-Control-Allow-Headers", "Content-Type"); w.Header().Set("Access-Control-Allow-Methods", "GET,POST,PUT,OPTIONS")
	if r.Method == http.MethodOptions { w.WriteHeader(http.StatusNoContent); return }
	if r.URL.Path == "/healthz" { writeJSON(w, http.StatusOK, map[string]string{"status":"ok"}); return }
	path := strings.Trim(r.URL.Path,"/"); parts := strings.Split(path,"/")
	if len(parts)==2 && parts[0]=="api" && parts[1]=="functions" && r.Method==http.MethodGet { h.functions(w,r); return }
	if len(parts)==2 && parts[0]=="api" && parts[1]=="projects" && r.Method==http.MethodGet { h.projects(w,r); return }
	if len(parts)==2 && parts[0]=="api" && parts[1]=="projects" && r.Method==http.MethodPost { h.createProject(w,r); return }
	if len(parts)>=3 && parts[0]=="api" && parts[1]=="projects" {
		id := parts[2]
		if len(parts)==3 && r.Method==http.MethodGet { h.project(w,r,id); return }
		if len(parts)==4 && parts[3]=="profile" && r.Method==http.MethodGet { h.profile(w,r,id); return }
		if len(parts)==4 && parts[3]=="summary" && r.Method==http.MethodGet { h.summary(w,r,id); return }
		if len(parts)==5 && parts[3]=="profile" && r.Method==http.MethodPut { h.updateProfile(w,r,id,parts[4]); return }
	}
	writeError(w,http.StatusNotFound,"not_found","route not found")
}

func (h *Handler) functions(w http.ResponseWriter,r *http.Request) { data,err:=h.Store.ListFunctions(r.Context()); if err!=nil { writeError(w,500,"internal_error","could not load catalog"); return }; writeJSON(w,200,data) }
func (h *Handler) projects(w http.ResponseWriter,r *http.Request) { data,err:=h.Store.ListProjects(r.Context()); if err!=nil { writeError(w,500,"internal_error","could not load projects"); return }; writeJSON(w,200,data) }
func (h *Handler) createProject(w http.ResponseWriter,r *http.Request) { var input struct{Name string `json:"name"`; OrganizationName string `json:"organizationName"`}; if err:=decodeJSON(r,&input); err!=nil { writeError(w,400,"invalid_json",err.Error()); return }; if strings.TrimSpace(input.Name)=="" { writeError(w,400,"validation_error","project name is required"); return }; if strings.TrimSpace(input.OrganizationName)=="" { input.OrganizationName="Unnamed organization" }; p,err:=h.Store.CreateProject(r.Context(),input.Name,input.OrganizationName); if err!=nil { writeError(w,500,"internal_error","could not create project"); return }; writeJSON(w,201,p) }
func (h *Handler) project(w http.ResponseWriter,r *http.Request,id string) { p,err:=h.Store.GetProject(r.Context(),id); if err!=nil { writeError(w,404,"not_found","project not found"); return }; writeJSON(w,200,p) }
func (h *Handler) profile(w http.ResponseWriter,r *http.Request,id string) { p,err:=h.Store.ListProfile(r.Context(),id); if err!=nil { writeError(w,500,"internal_error","could not load profile"); return }; writeJSON(w,200,p) }
func (h *Handler) updateProfile(w http.ResponseWriter,r *http.Request,projectID,subcategoryID string) { var patch store.ProfilePatch; if err:=decodeJSON(r,&patch); err!=nil { writeError(w,400,"invalid_json",err.Error()); return }; for _,level:=range []*string{patch.CurrentCoverageLevel,patch.TargetCoverageLevel} { if level!=nil { if _,err:=domain.Score(domain.CoverageLevel(*level)); err!=nil { writeError(w,400,"validation_error",err.Error()); return } } }; p,err:=h.Store.UpdateProfile(r.Context(),projectID,subcategoryID,patch); if err!=nil { log.Printf("profile update failed: %v",err); if strings.Contains(err.Error(),"invalid") || strings.Contains(err.Error(),"no fields") { writeError(w,400,"validation_error",err.Error()) } else { writeError(w,404,"not_found","profile not found") }; return }; writeJSON(w,200,p) }
type FunctionSummary struct { Code string `json:"code"`; CoveragePct float64 `json:"coveragePct"`; IncludedCount int `json:"includedCount"` }
type summaryResponse struct { domain.Summary; Functions []FunctionSummary `json:"functions"` }
func (h *Handler) summary(w http.ResponseWriter,r *http.Request,id string) { rows,err:=h.Store.ListProfile(r.Context(),id); if err!=nil { writeError(w,500,"internal_error","could not calculate summary"); return }; all:=make([]domain.ProfileScore,0,len(rows)); groups:=map[string][]domain.ProfileScore{}; for _,row:=range rows { score:=domain.ProfileScore{Included:row.Included,Current:domain.CoverageLevel(row.CurrentCoverageLevel),Target:domain.CoverageLevel(row.TargetCoverageLevel)}; all=append(all,score); groups[row.FunctionCode]=append(groups[row.FunctionCode],score) }; out:=summaryResponse{Summary:domain.CalculateSummary(all)}; for code,items:=range groups { s:=domain.CalculateSummary(items); out.Functions=append(out.Functions,FunctionSummary{Code:code,CoveragePct:s.CoveragePct,IncludedCount:s.IncludedCount}) }; writeJSON(w,200,out) }

func decodeJSON(r *http.Request,v any) error { dec:=json.NewDecoder(r.Body); if err:=dec.Decode(v); err!=nil { return err }; if dec.More() { return errors.New("request must contain one JSON object") }; return nil }
func writeJSON(w http.ResponseWriter,status int,v any) { w.Header().Set("Content-Type","application/json"); w.WriteHeader(status); _=json.NewEncoder(w).Encode(v) }
func writeError(w http.ResponseWriter,status int,code,message string) { var body errorBody; body.Error.Code=code; body.Error.Message=message; writeJSON(w,status,body) }
