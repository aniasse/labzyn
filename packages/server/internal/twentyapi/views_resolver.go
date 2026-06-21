package twentyapi

import (
	"context"
	"time"
)

type viewT struct {
	Id                                      UUID
	Name                                    string
	ObjectMetadataId                        UUID
	Type                                    string
	Key                                     *string
	Icon                                    *string
	Position                                float64
	IsCompact                               bool
	OpenRecordIn                            *string
	KanbanAggregateOperation                *string
	KanbanAggregateOperationFieldMetadataId *UUID
	MainGroupByFieldMetadataId              *UUID
	ShouldHideEmptyGroups                   bool
	AnyFieldFilterValue                     *string
	CalendarFieldMetadataId                 *UUID
	CalendarLayout                          *string
	Visibility                              *string
	CreatedByUserWorkspaceId                *UUID
	IsActive                                bool
	ViewFields                              []*viewFieldT
	ViewFieldGroups                         []*viewFieldGroupT
	ViewFilters                             []*viewFilterT
	ViewFilterGroups                        []*viewFilterGroupT
	ViewSorts                               []*viewSortT
	ViewGroups                              []*viewGroupT
}

type viewFieldT struct {
	Id                 UUID
	FieldMetadataId    UUID
	ViewId             UUID
	IsVisible          bool
	Position           float64
	Size               float64
	AggregateOperation *string
	ViewFieldGroupId   *UUID
	IsActive           bool
	CreatedAt          DateTime
	UpdatedAt          DateTime
	DeletedAt          *DateTime
}

type viewFieldGroupT struct {
	Id         UUID
	Name       string
	Position   float64
	IsVisible  bool
	ViewId     UUID
	IsActive   bool
	CreatedAt  DateTime
	UpdatedAt  DateTime
	DeletedAt  *DateTime
	ViewFields []*viewFieldT
}

type viewFilterT struct {
	Id                            UUID
	FieldMetadataId               UUID
	Operand                       string
	Value                         *JSON
	ViewFilterGroupId             *UUID
	PositionInViewFilterGroup     *float64
	SubFieldName                  *string
	RelationTargetFieldMetadataId *UUID
	ViewId                        UUID
	CreatedAt                     DateTime
	UpdatedAt                     DateTime
	DeletedAt                     *DateTime
}

type viewFilterGroupT struct {
	Id                        UUID
	ParentViewFilterGroupId   *UUID
	LogicalOperator           string
	PositionInViewFilterGroup *float64
	ViewId                    UUID
}

type viewSortT struct {
	Id              UUID
	FieldMetadataId UUID
	Direction       string
	SubFieldName    *string
	ViewId          UUID
	CreatedAt       DateTime
	DeletedAt       *DateTime
	UpdatedAt       DateTime
}

type viewGroupT struct {
	Id         UUID
	IsVisible  bool
	FieldValue string
	Position   float64
	ViewId     UUID
	CreatedAt  DateTime
	UpdatedAt  DateTime
	DeletedAt  *DateTime
}

// GetViews returns one TABLE view per object, with a column per field. The
// frontend uses these to render the record table.
func (r *MetadataResolver) GetViews(ctx context.Context, args struct {
	ObjectMetadataId *string
	ViewTypes        *[]string
}) ([]*viewT, error) {
	// Only TABLE views are synthesized; widget-type requests get nothing.
	if args.ViewTypes != nil {
		wantsTable := false
		for _, t := range *args.ViewTypes {
			if t == "TABLE" {
				wantsTable = true
			}
		}
		if !wantsTable {
			return []*viewT{}, nil
		}
	}

	wsID, err := r.core.DefaultWorkspaceID(ctx, userIDFromCtx(ctx))
	if err != nil {
		return nil, err
	}
	objs, err := r.meta.ListObjects(ctx, wsID)
	if err != nil {
		return nil, err
	}

	now := DateTime{time.Now()}
	views := []*viewT{}
	for _, o := range objs {
		if args.ObjectMetadataId != nil && *args.ObjectMetadataId != o.ID {
			continue
		}
		fields, err := r.meta.ListFields(ctx, o.ID)
		if err != nil {
			return nil, err
		}
		viewID := synthFieldID(o.ID, "view")

		// One column per field, in the same id order as the objects query.
		fieldIDs := []UUID{
			synthFieldID(o.ID, "id"),
			synthFieldID(o.ID, "createdAt"),
			synthFieldID(o.ID, "updatedAt"),
		}
		for _, f := range fields {
			fieldIDs = append(fieldIDs, UUID(f.ID))
		}

		viewFields := make([]*viewFieldT, 0, len(fieldIDs))
		for i, fid := range fieldIDs {
			viewFields = append(viewFields, &viewFieldT{
				Id:              synthFieldID(string(viewID), string(fid)),
				FieldMetadataId: fid,
				ViewId:          viewID,
				IsVisible:       true,
				Position:        float64(i),
				Size:            180,
				IsActive:        true,
				CreatedAt:       now,
				UpdatedAt:       now,
			})
		}

		key := "INDEX"
		views = append(views, &viewT{
			Id:                    viewID,
			Name:                  "All " + o.LabelPlural,
			ObjectMetadataId:      UUID(o.ID),
			Type:                  "TABLE",
			Key:                   &key,
			Position:              0,
			IsCompact:             false,
			ShouldHideEmptyGroups: false,
			IsActive:              true,
			ViewFields:            viewFields,
			ViewFieldGroups:       []*viewFieldGroupT{},
			ViewFilters:           []*viewFilterT{},
			ViewFilterGroups:      []*viewFilterGroupT{},
			ViewSorts:             []*viewSortT{},
			ViewGroups:            []*viewGroupT{},
		})
	}
	return views, nil
}
