package namer_client

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/3ideas/psasim/lib/compdb"
	pb "github.com/3ideas/psasim/lib/namer_service" // Adjust import path as necessary
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type NameClient struct {
	conn   *grpc.ClientConn
	client pb.NamerServiceClient
}

func Connect() (*NameClient, error) {
	// conn, err := grpc.Dial("server_address:port", grpc.WithInsecure()) // Deprecated
	client, err := grpc.NewClient("127.0.0.1:50051", grpc.WithTransportCredentials(insecure.NewCredentials())) // Use NewClient instead

	if err != nil {
		return nil, fmt.Errorf("unable to connect to server. Has it been started? %v", err)
	}
	slog.Info("Connected to name server")
	nameClient := &NameClient{conn: client}
	nameClient.client = pb.NewNamerServiceClient(client)

	return nameClient, nil
}

func (c *NameClient) Close() {
	c.conn.Close()
}

func (c *NameClient) GetName(alias string) (*compdb.NameDetails, error) {
	response, err := c.client.GetName(context.Background(), &pb.ComponentAlias{Alias: alias})
	if err != nil {
		return nil, fmt.Errorf("could not get name: %v", err)
	}
	if response.Error != "" {
		return nil, fmt.Errorf("could not get NameDetails for :%s - %s", alias, response.Error)
	}

	return c.convertNameResponse(response), nil
}

func (c *NameClient) GetHierarchyByAlias(alias string) (compdb.Hierarchy, error) {
	response, err := c.client.GetHierarchyByAlias(context.Background(), &pb.GetHierarchyByAliasRequest{Alias: alias})
	if err != nil {
		return nil, fmt.Errorf("could not get hierarchy: %v", err)
	}

	hierarchy := compdb.Hierarchy{}
	for _, comp := range response.Hierarchy {
		hierarchy = append(hierarchy, c.convertComponentInfo(comp))
	}

	return hierarchy, nil
}

func (c *NameClient) convertNameResponse(name *pb.GetNameResponse) *compdb.NameDetails {
	return &compdb.NameDetails{
		Name:     name.Name,
		RuleName: name.NameRule,
		Location: c.convertNamePartResponse(name.Location),
		Circuit:  c.convertNamePartResponse(name.Circuit),
		Plant:    c.convertNamePartResponse(name.Plant),
		Origin:   c.convertNamePartResponse(name.Origin),
		Pathname: name.Pathname,
		Alias:    name.Alias,
	}
}

func (c *NameClient) GetNameWithHierarchy(alias string) (*compdb.NameWithHierachy, error) {
	response, err := c.client.GetNameWithHierarchy(context.Background(), &pb.ComponentAlias{Alias: alias})
	if err != nil {
		return nil, fmt.Errorf("could not get name with hierarchy: %v", err)
	}
	if response.Error != "" {
		return nil, fmt.Errorf("could not get NameWithHierachy for :%s - %s", alias, response.Error)
	}

	namePart := c.convertNameResponse(response.Name)

	parents := []*compdb.ComponentInfo{}
	for _, comp := range response.Hierarchy {
		parents = append(parents, c.convertComponentInfo(comp))
	}

	return &compdb.NameWithHierachy{NameDetails: *namePart, Hierarchy: parents}, nil
}

func (c *NameClient) convertComponentInfo(comp *pb.ComponentInfo) *compdb.ComponentInfo {
	return &compdb.ComponentInfo{
		Alias:               comp.Alias,
		Path:                comp.Path,
		ComponentClassName:  comp.ComponentClassname,
		SubstationClassName: comp.SubstationClassname,
		ID:                  comp.Id,
		CloneID:             comp.CloneID,
		ClonePathname:       comp.ClonePathname,
		NameRule:            comp.NameRule,
		InSymbol:            comp.InSymbol,
	}
}

func (c *NameClient) convertNamePartResponse(namePart *pb.NamePartResponse) *compdb.NamePartResponse {
	namePartDetails := []*compdb.NamePartDetailsResponse{}
	for _, detail := range namePart.NamePartDetails {
		namePartDetails = append(namePartDetails, &compdb.NamePartDetailsResponse{Value: detail.Value, Source: detail.Source, RawValue: detail.RawValue, Separator: detail.Separator})
	}
	return &compdb.NamePartResponse{Value: namePart.Value, Used: namePart.Used, Alias: namePart.Alias, Details: namePartDetails}
}

func (c *NameClient) RenameComponent(alias, newName string) error {
	response, err := c.client.RenameComponent(context.Background(), &pb.RenameComponentRequest{Alias: alias, NewName: newName})
	if err != nil {
		return fmt.Errorf("could not rename: %v", err)
	}
	if response.Error != "" {
		return fmt.Errorf("could not rename: %v", response.Error)
	}
	return nil
}

func (c *NameClient) MoveComponent(alias, newLocationAlias string) error {
	response, err := c.client.MoveComponent(context.Background(), &pb.MoveComponentRequest{Alias: alias, NewLocationAlias: newLocationAlias})
	if err != nil {
		return fmt.Errorf("could not move: %v", err)
	}
	if response.Error != "" {
		return fmt.Errorf("could not move: %v", response.Error)
	}
	return nil
}

func (c *NameClient) CreateAttribute(alias, attrName, attrValue string) error {
	response, err := c.client.CreateAttribute(context.Background(), &pb.CreateAttributeRequest{Alias: alias, AttrName: attrName, AttrValue: attrValue})
	if err != nil {
		return fmt.Errorf("could not create attribute: %v", err)
	}
	if response.Error != "" {
		return fmt.Errorf("could not create attribute: %v", response.Error)
	}
	return nil
}

func (c *NameClient) UpdateAttribute(alias, attrName, attrValue string) error {
	response, err := c.client.UpdateAttribute(context.Background(), &pb.UpdateAttributeRequest{Alias: alias, AttrName: attrName, AttrValue: attrValue})
	if err != nil {
		return fmt.Errorf("could not update attribute: %v", err)
	}
	if response.Error != "" {
		return fmt.Errorf("could not update attribute: %v", response.Error)
	}
	return nil
}

func (c *NameClient) CreateComponent(alias, name, parentAlias, templateAlias, substationClassName string) error {
	response, err := c.client.CreateComponent(context.Background(), &pb.CreateComponentRequest{
		Alias:               alias,
		Name:                name,
		ParentAlias:         parentAlias,
		TemplateAlias:       templateAlias,
		SubstationClassName: substationClassName,
	})
	if err != nil {
		return fmt.Errorf("could not create new component: %v", err)
	}
	if response.Error != "" {
		return fmt.Errorf("could not create new component: %v", response.Error)
	}
	return nil
}

func (c *NameClient) RollbackAll() error {
	response, err := c.client.RollbackAll(context.Background(), &pb.RollbackAllRequest{})
	if err != nil {
		return fmt.Errorf("could not rollback all: %v", err)
	}
	if response.Error != "" {
		return fmt.Errorf("could not rollback all: %v", response.Error)
	}
	return nil
}

func (c *NameClient) GetNumberOfChanges() (int, error) {
	response, err := c.client.GetNumberOfChanges(context.Background(), &pb.GetNumberOfChangesRequest{})
	if err != nil {
		return 0, fmt.Errorf("could not get number of changes: %v", err)
	}
	return int(response.NumberOfChanges), nil
}

func (c *NameClient) SetRollbackPoint() error {
	slog.Info("Setting rollback point")
	response, err := c.client.SetRollbackPoint(context.Background(), &pb.SetRollbackPointRequest{})
	if err != nil {
		return fmt.Errorf("could not set rollback point: %v", err)
	}
	if response.Error != "" {
		return fmt.Errorf("could not set rollback point: %v", response.Error)
	}
	return nil
}

func (c *NameClient) RollbackToPoint() error {
	slog.Info("Rolling back to point")
	response, err := c.client.RollbackToPoint(context.Background(), &pb.RollbackToPointRequest{})
	if err != nil {
		return fmt.Errorf("could not rollback to point: %v", err)
	}
	if response.Error != "" {
		return fmt.Errorf("could not rollback to point: %v", response.Error)
	}
	return nil
}

func (c *NameClient) GetAttributeValue(alias, attrName string) (compdb.AttributeValue, error) {
	response, err := c.client.GetAttributeValue(context.Background(), &pb.GetAttributeValueRequest{Alias: alias, AttrName: attrName})
	if err != nil {
		return compdb.AttributeValue{}, fmt.Errorf("could not get attribute: %v", err)
	}
	if response.Error != "" {
		return compdb.AttributeValue{}, fmt.Errorf("could not get attribute: %v", response.Error)
	}
	return compdb.AttributeValue{
		Name:       attrName,
		Value:      response.AttrValue,
		ID:         response.Id,
		CompID:     response.CompId,
		Definition: response.Definition,
	}, nil
}

func (c *NameClient) GetComponentClassDetails(alias string) (*compdb.ComponentClassDetails, error) {
	response, err := c.client.GetComponentClass(context.Background(), &pb.GetComponentClassRequest{Alias: alias})
	if err != nil {
		return nil, fmt.Errorf("could not get component class: %v", err)
	}
	if response.Error != "" {
		return nil, fmt.Errorf("could not get component class: %v", response.Error)
	}
	return &compdb.ComponentClassDetails{ClassName: response.ComponentClassName, NameRule: response.ComponentClassNameRule, SubstationClass: compdb.SubstationType(response.SubstationClass)}, nil
}

func (c *NameClient) GetComponentInfoByID(id string) (*compdb.ComponentInfo, error) {
	response, err := c.client.GetComponentByID(context.Background(), &pb.ComponentID{ComponentID: id})
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, fmt.Errorf("could not get component info: %v", response.Error)
	}
	return c.convertComponentInfo(response.CompInfo), nil
}

func (c *NameClient) GetChildrenInfoByID(id string) ([]*compdb.ComponentInfo, error) {
	response, err := c.client.GetChildrenInfoByID(context.Background(), &pb.ComponentID{ComponentID: id})
	if err != nil {
		return nil, err
	}
	if response.Error != "" {
		return nil, fmt.Errorf("could not get children info: %v", response.Error)
	}

	children := []*compdb.ComponentInfo{}
	for _, child := range response.Children {
		children = append(children, c.convertComponentInfo(child))
	}
	return children, nil
}

func (c *NameClient) GetComponentInfo(alias string) (*compdb.ComponentInfo, error) {
	response, err := c.client.GetComponentInfo(context.Background(), &pb.ComponentAlias{Alias: alias})
	if response.Error != "" {
		return nil, fmt.Errorf("could not get component info: %v", response.Error)
	}

	if err != nil {
		return nil, err
	}
	return c.convertComponentInfo(response.CompInfo), nil
}
