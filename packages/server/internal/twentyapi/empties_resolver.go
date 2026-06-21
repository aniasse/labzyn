package twentyapi

import (
	"context"
)

// Queries the frontend calls at boot that we don't model yet; returning empty
// lists lets the app finish rendering (these power optional widgets/menus).

type frontComponentT struct {
	Id                     UUID
	Name                   string
	ApplicationId          *string
	BuiltComponentChecksum *string
	BuiltComponentPath     *string
	ComponentName          *string
	CreatedAt              DateTime
	Description            *string
	IsHeadless             bool
	SourceComponentPath    *string
	UniversalIdentifier    *string
	UpdatedAt              DateTime
	UsesSdkClient          bool
}

type logicFunctionT struct {
	Id                            UUID
	Name                          string
	Description                   *string
	Runtime                       *string
	TimeoutSeconds                *float64
	ExecutionMode                 *string
	SourceHandlerPath             *string
	HandlerName                   *string
	CronTriggerSettings           *JSON
	DatabaseEventTriggerSettings  *JSON
	HttpRouteTriggerSettings      *JSON
	ToolTriggerSettings           *JSON
	WorkflowActionTriggerSettings *JSON
	ApplicationId                 *string
	UniversalIdentifier           *string
	CreatedAt                     DateTime
	UpdatedAt                     DateTime
}

type navigationMenuItemT struct {
	Id                     UUID
	Type                   string
	UserWorkspaceId        *UUID
	TargetRecordId         *UUID
	TargetObjectMetadataId *UUID
	ViewId                 *UUID
	FolderId               *UUID
	Name                   *string
	Link                   *string
	Icon                   *string
	Color                  *string
	PageLayoutId           *UUID
	Position               *float64
	ApplicationId          *string
	CreatedAt              DateTime
	UpdatedAt              DateTime
	TargetRecordIdentifier *navTargetRecordIdentifierT
}

type navTargetRecordIdentifierT struct {
	Id              UUID
	LabelIdentifier *string
	ImageIdentifier *string
}

func (r *MetadataResolver) FrontComponents(_ context.Context) ([]*frontComponentT, error) {
	return []*frontComponentT{}, nil
}

func (r *MetadataResolver) FindManyLogicFunctions(_ context.Context) ([]*logicFunctionT, error) {
	return []*logicFunctionT{}, nil
}

func (r *MetadataResolver) NavigationMenuItems(_ context.Context) ([]*navigationMenuItemT, error) {
	return []*navigationMenuItemT{}, nil
}
