package namer_server

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"log/slog"
	"net"
	"net/http"

	"github.com/gorilla/mux" // Make sure to import the gorilla/mux package

	"github.com/3ideas/psasim/lib/compdb"
	pb "github.com/3ideas/psasim/lib/namer_service" // Adjust import path as necessary

	"google.golang.org/grpc"
)

type server struct {
	pb.UnimplementedNamerServiceServer
	namer *compdb.ComponentDb
}

func NewNameServer(namer *compdb.ComponentDb) *server {
	return &server{
		namer: namer,
	}
}

func (s *server) GetName(ctx context.Context, req *pb.ComponentAlias) (*pb.GetNameResponse, error) {
	nameDetails, err := s.namer.GetNameFull(req.Alias)
	if err != nil {
		slog.Warn("Failed to resolve name", "alias", req.Alias, "error", err)
		return &pb.GetNameResponse{
			Name:  "",
			Error: err.Error(),
		}, nil
	}
	slog.Info("Resolved name", "alias", req.Alias, "name", nameDetails.Name, "rule", nameDetails.Rule, "location", nameDetails.Location, "circuit", nameDetails.Circuit, "plant", nameDetails.Plant, "origin", nameDetails.Origin)

	nameResponse := convertNameDetails(&nameDetails.NameDetails)

	for i := len(nameDetails.Parents) - 1; i >= 0; i-- {
		comp := nameDetails.Parents[i]
		slog.Debug("Parents", "alias", comp.ComponentAlias, "name", comp.ComponentPathname, "SubstationClass", comp.ComponentSubstationClass, "Class", comp.ComponentClass, "Parent", comp.ComponentParentID, "ID", comp.ComponentID)
	}

	return nameResponse, nil
}

func (s *server) GetHierarchyByAlias(ctx context.Context, req *pb.GetHierarchyByAliasRequest) (*pb.GetHierarchyByAliasResponse, error) {
	hierarchy, err := s.namer.GetHierarchyByAlias(req.Alias)
	if err != nil {
		return &pb.GetHierarchyByAliasResponse{Error: err.Error()}, nil
	}
	return &pb.GetHierarchyByAliasResponse{Hierarchy: convertComponentInfoList(hierarchy)}, nil
}

func convertNameDetails(nameDetails *compdb.NameDetails) *pb.GetNameResponse {
	location := convertNamePartDetails(nameDetails.Location)
	circuit := convertNamePartDetails(nameDetails.Circuit)
	plant := convertNamePartDetails(nameDetails.Plant)
	origin := convertNamePartDetails(nameDetails.Origin)

	// for i := len(nameDetails.Parents) - 1; i >= 0; i-- {
	// 	comp := nameDetails.Parents[i]
	// 	slog.Debug("Parents", "alias", comp.ComponentAlias, "name", comp.ComponentPathname, "SubstationClass", comp.ComponentSubstationClass, "Class", comp.ComponentClass, "Parent", comp.ComponentParentID, "ID", comp.ComponentID)
	// }

	return &pb.GetNameResponse{
		Name:     nameDetails.Name,
		Location: location,
		Circuit:  circuit,
		Plant:    plant,
		Origin:   origin,
		NameRule: nameDetails.RuleName,
		Error:    "",
		Pathname: nameDetails.Pathname,
		Alias:    nameDetails.Alias,
	}
}

func (s *server) GetNameWithHierarchy(ctx context.Context, req *pb.ComponentAlias) (*pb.GetNameWithHierarchyResponse, error) {
	nameDetails, err := s.namer.GetNameWithHierarchy(req.Alias)
	if err != nil {
		slog.Warn("Failed to resolve name", "alias", req.Alias, "error", err)
		return &pb.GetNameWithHierarchyResponse{
			Error: err.Error(),
		}, nil
	}
	nameResponse := convertNameDetails(&nameDetails.NameDetails)
	if err != nil {
		return &pb.GetNameWithHierarchyResponse{Error: err.Error()}, nil
	}

	return &pb.GetNameWithHierarchyResponse{Name: nameResponse, Hierarchy: convertComponentInfoList(nameDetails.Hierarchy)}, nil
}

func convertNamePartDetails(part *compdb.NamePartResponse) *pb.NamePartResponse {
	namePartDetails := []*pb.NamePartDetailResponse{}
	for _, detail := range part.Details {
		namePartDetails = append(namePartDetails, convertNamePartDetailsResponse(detail))
	}
	namePartResponse := &pb.NamePartResponse{Value: part.Value, Used: part.Used, Alias: part.Alias, NamePartDetails: namePartDetails}
	return namePartResponse
}

func convertNamePartDetailsResponse(detail *compdb.NamePartDetailsResponse) *pb.NamePartDetailResponse {
	return &pb.NamePartDetailResponse{Value: detail.Value, Source: detail.Source, RawValue: detail.RawValue, Separator: detail.Separator}
}

// Rename method implementation
func (s *server) RenameComponent(ctx context.Context, req *pb.RenameComponentRequest) (*pb.RenameComponentResponse, error) {
	err := s.namer.RenameComponent(req.Alias, req.NewName)
	if err != nil {
		return &pb.RenameComponentResponse{Error: err.Error()}, nil
	}
	return &pb.RenameComponentResponse{Error: ""}, nil
}

// Move method implementation
func (s *server) MoveComponent(ctx context.Context, req *pb.MoveComponentRequest) (*pb.MoveComponentResponse, error) {
	err := s.namer.MoveComponent(req.Alias, req.NewLocationAlias)
	if err != nil {
		return &pb.MoveComponentResponse{Error: err.Error()}, nil
	}
	return &pb.MoveComponentResponse{Error: ""}, nil
}

// CreateAttribute method implementation
func (s *server) CreateAttribute(ctx context.Context, req *pb.CreateAttributeRequest) (*pb.CreateAttributeResponse, error) {
	err := s.namer.CreateAttribute(req.Alias, req.AttrName, req.AttrValue)
	if err != nil {
		return &pb.CreateAttributeResponse{Error: err.Error()}, nil
	}
	return &pb.CreateAttributeResponse{Error: ""}, nil
}

// UpdateAttribute method implementation
func (s *server) UpdateAttribute(ctx context.Context, req *pb.UpdateAttributeRequest) (*pb.UpdateAttributeResponse, error) {
	err := s.namer.UpdateAttribute(req.Alias, req.AttrName, req.AttrValue)
	if err != nil {
		return &pb.UpdateAttributeResponse{Error: err.Error()}, nil
	}
	return &pb.UpdateAttributeResponse{Error: ""}, nil
}

// CreateNewComp method implementation
func (s *server) CreateComponent(ctx context.Context, req *pb.CreateComponentRequest) (*pb.CreateComponentResponse, error) {
	err := s.namer.CreateComponent(req.Alias, req.Name, req.ParentAlias, req.TemplateAlias, req.SubstationClassName)
	if err != nil {
		return &pb.CreateComponentResponse{Error: err.Error()}, nil
	}
	return &pb.CreateComponentResponse{Error: ""}, nil
}

// RollbackAll method implementation
func (s *server) RollbackAll(ctx context.Context, req *pb.RollbackAllRequest) (*pb.RollbackAllResponse, error) {
	err := s.namer.RollbackAll()
	if err != nil {
		return &pb.RollbackAllResponse{Error: err.Error()}, nil
	}
	return &pb.RollbackAllResponse{Error: ""}, nil
}

// GetNumberOfChanges method implementation
func (s *server) GetNumberOfChanges(ctx context.Context, req *pb.GetNumberOfChangesRequest) (*pb.GetNumberOfChangesResponse, error) {
	numChanges, _ := s.namer.GetNumberOfChanges()
	return &pb.GetNumberOfChangesResponse{NumberOfChanges: int32(numChanges)}, nil
}

// Add this function to handle the JSON request
func (s *server) GetNameJSON(w http.ResponseWriter, r *http.Request) {
	var req pb.ComponentAlias
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the existing gRPC method
	resp, err := s.GetName(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response as JSON
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, _ := json.MarshalIndent(resp, "", "  ") // Indent with 2 spaces
	w.Write(jsonResponse)
}

func (s *server) GetAttributeValue(ctx context.Context, req *pb.GetAttributeValueRequest) (*pb.GetAttributeValueResponse, error) {
	attr, err := s.namer.GetAttributeValue(req.Alias, req.AttrName)
	if err != nil {
		return &pb.GetAttributeValueResponse{Error: err.Error()}, nil
	}
	return &pb.GetAttributeValueResponse{AttrValue: attr.Value, Id: attr.ID, CompId: attr.CompID, Definition: attr.Definition, Error: ""}, nil
}

// Add this function to handle the JSON request for GetAttributeValue
func (s *server) GetAttributeValueJSON(w http.ResponseWriter, r *http.Request) {
	var req pb.GetAttributeValueRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	// Call the existing gRPC method
	resp, err := s.GetAttributeValue(context.Background(), &req)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the response as nicely formatted JSON
	w.Header().Set("Content-Type", "application/json")
	jsonResponse, _ := json.MarshalIndent(resp, "", "  ") // Indent with 2 spaces
	w.Write(jsonResponse)
}

func (s *server) GetComponentClass(ctx context.Context, req *pb.GetComponentClassRequest) (*pb.GetComponentClassResponse, error) {

	classDetails, err := s.namer.GetComponentClassDetails(req.Alias)
	if err != nil {
		return &pb.GetComponentClassResponse{Error: fmt.Errorf("error getting component class definition: %w", err).Error()}, nil
	}
	return &pb.GetComponentClassResponse{ComponentClassName: classDetails.ClassName, ComponentClassNameRule: classDetails.NameRule, SubstationClass: int32(classDetails.SubstationClass), Error: ""}, nil
}

// SetRollbackPoint method implementation
func (s *server) SetRollbackPoint(ctx context.Context, req *pb.SetRollbackPointRequest) (*pb.SetRollbackPointResponse, error) {
	s.namer.SetRollbackPoint()
	return &pb.SetRollbackPointResponse{Error: ""}, nil
}

// RollbackToPoint method implementation
func (s *server) RollbackToPoint(ctx context.Context, req *pb.RollbackToPointRequest) (*pb.RollbackToPointResponse, error) {
	err := s.namer.RollbackToPoint()
	if err != nil {
		return &pb.RollbackToPointResponse{Error: err.Error()}, nil
	}
	return &pb.RollbackToPointResponse{Error: ""}, nil
}

func (s *server) GetComponentInfoByID(ctx context.Context, req *pb.ComponentID) (*pb.ComponentInfoResponse, error) {
	compInfo, err := s.namer.GetComponentInfoByID(req.ComponentID)
	if err != nil {
		return &pb.ComponentInfoResponse{Error: err.Error()}, nil
	}
	c := convertComponentInfo(compInfo)

	compInfoResponse := &pb.ComponentInfoResponse{
		CompInfo: c,
	}

	return compInfoResponse, nil
}

func (s *server) GetComponentInfo(ctx context.Context, req *pb.ComponentAlias) (*pb.ComponentInfoResponse, error) {

	compInfo, err := s.namer.GetComponentInfo(req.Alias)
	if err != nil {
		return &pb.ComponentInfoResponse{Error: err.Error()}, nil
	}
	c := convertComponentInfo(compInfo)

	compInfoResponse := &pb.ComponentInfoResponse{
		CompInfo: c,
	}

	return compInfoResponse, nil
}

func (s *server) GetChildrenInfoByID(ctx context.Context, req *pb.ComponentID) (*pb.GetChildrenByIDResponse, error) {
	children, err := s.namer.GetChildrenInfoByID(req.ComponentID)
	if err != nil {
		return &pb.GetChildrenByIDResponse{Error: "No children found"}, nil
	}

	childrenResponse := convertComponentInfoList(children)

	return &pb.GetChildrenByIDResponse{Children: childrenResponse}, nil
}

func convertComponentInfo(compInfo *compdb.ComponentInfo) *pb.ComponentInfo {
	childInfoR := &pb.ComponentInfo{
		Alias:               compInfo.Alias,
		Path:                compInfo.Path,
		ComponentClassname:  compInfo.ComponentClassName,
		SubstationClassname: compInfo.SubstationClassName,
		Id:                  compInfo.ID,
		CloneID:             compInfo.CloneID,
		ClonePathname:       compInfo.ClonePathname,
		NameRule:            compInfo.NameRule,
		InSymbol:            compInfo.InSymbol,
	}
	return childInfoR
}

func convertComponentInfoList(compInfoList []*compdb.ComponentInfo) []*pb.ComponentInfo {
	componentInfo := []*pb.ComponentInfo{}
	for _, parent := range compInfoList {
		componentInfo = append(componentInfo, convertComponentInfo(parent))
	}
	return componentInfo
}

// Modify the StartServer method to include the new HTTP handler
func (s *server) StartServer() error {
	lis, err := net.Listen("tcp", "127.0.0.1:50051")
	if err != nil {
		slog.Error("Failed to open listener", "error", err)
		log.Fatalf("failed to listen: %v", err)
	}

	grpcServer := grpc.NewServer()
	pb.RegisterNamerServiceServer(grpcServer, s)

	// Set up the HTTP server
	router := mux.NewRouter()
	router.HandleFunc("/getname", s.GetNameJSON).Methods("POST")
	router.HandleFunc("/getattributevalue", s.GetAttributeValueJSON).Methods("POST") // Add this line

	go func() {
		if err := http.ListenAndServe(":50052", router); err != nil {
			slog.Error("Failed to start HTTP server", "error", err)
			log.Fatalf("failed to serve HTTP: %v", err)
		}
	}()

	slog.Info("Name server started")

	if err := grpcServer.Serve(lis); err != nil {
		slog.Error("Failed to start server", "error", err)
		log.Fatalf("failed to serve: %v", err)
	}

	return nil
}
