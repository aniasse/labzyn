package twentyapi

import (
	"context"
	"time"
)

type widgetGridPositionT struct {
	Column     *float64
	Row        *float64
	ColumnSpan *float64
	RowSpan    *float64
}

type ratioAggregateConfigT struct {
	FieldMetadataId *string
	OptionValue     *string
}

type standaloneRichTextBodyT struct {
	Blocknote *string
	Markdown  *string
}

type pageLayoutWidgetGridPositionT struct {
	LayoutMode *string
	Row        *string
	Column     *string
	RowSpan    *string
	ColumnSpan *string
}

type pageLayoutWidgetVerticalListPositionT struct {
	LayoutMode *string
	Index      *string
}

type pageLayoutWidgetCanvasPositionT struct {
	LayoutMode *string
}

type widgetPositionT struct {
	vertical *pageLayoutWidgetVerticalListPositionT
}

func (*widgetPositionT) ToPageLayoutWidgetGridPosition() (*pageLayoutWidgetGridPositionT, bool) {
	return nil, false
}

func (p *widgetPositionT) ToPageLayoutWidgetVerticalListPosition() (*pageLayoutWidgetVerticalListPositionT, bool) {
	return p.vertical, p.vertical != nil
}

func (*widgetPositionT) ToPageLayoutWidgetCanvasPosition() (*pageLayoutWidgetCanvasPositionT, bool) {
	return nil, false
}

type barChartConfigurationT struct {
	ConfigurationType                   *string
	AggregateFieldMetadataId            *string
	AggregateOperation                  *string
	PrimaryAxisGroupByFieldMetadataId   *string
	PrimaryAxisGroupBySubFieldName      *string
	PrimaryAxisDateGranularity          *string
	PrimaryAxisOrderBy                  *string
	PrimaryAxisManualSortOrder          *string
	SecondaryAxisGroupByFieldMetadataId *string
	SecondaryAxisGroupBySubFieldName    *string
	SecondaryAxisGroupByDateGranularity *string
	SecondaryAxisOrderBy                *string
	SecondaryAxisManualSortOrder        *string
	OmitNullValues                      *string
	AxisNameDisplay                     *string
	DisplayDataLabel                    *string
	DisplayLegend                       *string
	RangeMin                            *string
	RangeMax                            *string
	Color                               *string
	Description                         *string
	Filter                              *string
	GroupMode                           *string
	Layout                              *string
	IsCumulative                        *string
	SplitMultiValueFields               *string
	Timezone                            *string
	FirstDayOfTheWeek                   *string
}

type lineChartConfigurationT struct {
	ConfigurationType                   *string
	AggregateFieldMetadataId            *string
	AggregateOperation                  *string
	PrimaryAxisGroupByFieldMetadataId   *string
	PrimaryAxisGroupBySubFieldName      *string
	PrimaryAxisDateGranularity          *string
	PrimaryAxisOrderBy                  *string
	PrimaryAxisManualSortOrder          *string
	SecondaryAxisGroupByFieldMetadataId *string
	SecondaryAxisGroupBySubFieldName    *string
	SecondaryAxisGroupByDateGranularity *string
	SecondaryAxisOrderBy                *string
	SecondaryAxisManualSortOrder        *string
	OmitNullValues                      *string
	AxisNameDisplay                     *string
	DisplayDataLabel                    *string
	DisplayLegend                       *string
	RangeMin                            *string
	RangeMax                            *string
	Color                               *string
	Description                         *string
	Filter                              *string
	IsStacked                           *string
	IsCumulative                        *string
	SplitMultiValueFields               *string
	Timezone                            *string
	FirstDayOfTheWeek                   *string
}

type pieChartConfigurationT struct {
	ConfigurationType        *string
	GroupByFieldMetadataId   *string
	AggregateFieldMetadataId *string
	AggregateOperation       *string
	GroupBySubFieldName      *string
	DateGranularity          *string
	OrderBy                  *string
	ManualSortOrder          *string
	DisplayDataLabel         *string
	ShowCenterMetric         *string
	DisplayLegend            *string
	HideEmptyCategory        *string
	SplitMultiValueFields    *string
	Color                    *string
	Description              *string
	Filter                   *string
	Timezone                 *string
	FirstDayOfTheWeek        *string
}

type aggregateChartConfigurationT struct {
	ConfigurationType        *string
	AggregateFieldMetadataId *string
	AggregateOperation       *string
	Label                    *string
	DisplayDataLabel         *string
	NumberFormat             *string
	Description              *string
	Filter                   *string
	Prefix                   *string
	Suffix                   *string
	Timezone                 *string
	FirstDayOfTheWeek        *string
	RatioAggregateConfig     *ratioAggregateConfigT
}

type iframeConfigurationT struct {
	ConfigurationType *string
	Url               *string
}

type standaloneRichTextConfigurationT struct {
	ConfigurationType *string
	Body              *standaloneRichTextBodyT
}

type calendarConfigurationT struct {
	ConfigurationType *string
}

type emailsConfigurationT struct {
	ConfigurationType *string
}

type emailThreadConfigurationT struct {
	ConfigurationType *string
}

type fieldConfigurationT struct {
	ConfigurationType *string
	FieldDisplayMode  *string
	FieldMetadataId   *string
	ViewId            *string
}

type fieldRichTextConfigurationT struct {
	ConfigurationType *string
}

type fieldsConfigurationT struct {
	ConfigurationType                *string
	ViewId                           *string
	NewFieldDefaultVisibility        *string
	ShouldAllowUserToSeeHiddenFields *string
}

type filesConfigurationT struct {
	ConfigurationType *string
}

type notesConfigurationT struct {
	ConfigurationType *string
}

type tasksConfigurationT struct {
	ConfigurationType *string
}

type timelineConfigurationT struct {
	ConfigurationType *string
}

type viewConfigurationT struct {
	ConfigurationType *string
}

type recordTableConfigurationT struct {
	ConfigurationType *string
	ViewId            *string
}

type workflowConfigurationT struct {
	ConfigurationType *string
}

type workflowRunConfigurationT struct {
	ConfigurationType *string
}

type workflowVersionConfigurationT struct {
	ConfigurationType *string
}

type frontComponentConfigurationT struct {
	ConfigurationType *string
	FrontComponentId  *string
}

type widgetConfigurationT struct {
	fields *fieldsConfigurationT
}

func (*widgetConfigurationT) ToBarChartConfiguration() (*barChartConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToLineChartConfiguration() (*lineChartConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToPieChartConfiguration() (*pieChartConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToAggregateChartConfiguration() (*aggregateChartConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToIframeConfiguration() (*iframeConfigurationT, bool) { return nil, false }

func (*widgetConfigurationT) ToStandaloneRichTextConfiguration() (*standaloneRichTextConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToCalendarConfiguration() (*calendarConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToEmailsConfiguration() (*emailsConfigurationT, bool) { return nil, false }

func (*widgetConfigurationT) ToEmailThreadConfiguration() (*emailThreadConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToFieldConfiguration() (*fieldConfigurationT, bool) { return nil, false }

func (*widgetConfigurationT) ToFieldRichTextConfiguration() (*fieldRichTextConfigurationT, bool) {
	return nil, false
}

func (c *widgetConfigurationT) ToFieldsConfiguration() (*fieldsConfigurationT, bool) {
	return c.fields, c.fields != nil
}

func (*widgetConfigurationT) ToFilesConfiguration() (*filesConfigurationT, bool) { return nil, false }

func (*widgetConfigurationT) ToNotesConfiguration() (*notesConfigurationT, bool) { return nil, false }

func (*widgetConfigurationT) ToTasksConfiguration() (*tasksConfigurationT, bool) { return nil, false }

func (*widgetConfigurationT) ToTimelineConfiguration() (*timelineConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToViewConfiguration() (*viewConfigurationT, bool) { return nil, false }

func (*widgetConfigurationT) ToRecordTableConfiguration() (*recordTableConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToWorkflowConfiguration() (*workflowConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToWorkflowRunConfiguration() (*workflowRunConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToWorkflowVersionConfiguration() (*workflowVersionConfigurationT, bool) {
	return nil, false
}

func (*widgetConfigurationT) ToFrontComponentConfiguration() (*frontComponentConfigurationT, bool) {
	return nil, false
}

type pageLayoutWidgetT struct {
	Id                                UUID
	ApplicationId                     *string
	Title                             string
	Type                              string
	ObjectMetadataId                  *UUID
	CreatedAt                         DateTime
	UpdatedAt                         DateTime
	IsActive                          bool
	DeletedAt                         *DateTime
	ConditionalDisplay                *bool
	ConditionalAvailabilityExpression *string
	GridPosition                      widgetGridPositionT
	Position                          *widgetPositionT
	Configuration                     *widgetConfigurationT
	PageLayoutTabId                   UUID
}
type pageLayoutTabT struct {
	Id            UUID
	ApplicationId *string
	Title         string
	Icon          *string
	Position      float64
	LayoutMode    string
	Widgets       []*pageLayoutWidgetT
	PageLayoutId  UUID
	IsActive      bool
	CreatedAt     DateTime
	UpdatedAt     DateTime
}
type pageLayoutT struct {
	Id                                      UUID
	Name                                    string
	ObjectMetadataId                        *UUID
	Type                                    string
	UniversalIdentifier                     *string
	DefaultTabToFocusOnMobileAndSidePanelId *UUID
	CreatedAt                               DateTime
	UpdatedAt                               DateTime
	Tabs                                    []*pageLayoutTabT
}

// GetPageLayouts returns a record-page layout per object so the frontend uses
// our clean Fields layout instead of its default (whose Timeline/Tasks/Notes
// widgets are "invalid" for objects without activity support). The frontend
// matches a layout to an object by objectMetadataId.
func (r *MetadataResolver) GetPageLayouts(ctx context.Context, args struct{ PageLayoutType *string }) ([]*pageLayoutT, error) {
	if args.PageLayoutType != nil && *args.PageLayoutType != "RECORD_PAGE" {
		return []*pageLayoutT{}, nil
	}
	wsID, err := r.core.DefaultWorkspaceID(ctx, userIDFromCtx(ctx))
	if err != nil {
		return []*pageLayoutT{}, nil
	}
	objs, err := r.meta.ListObjects(ctx, wsID)
	if err != nil {
		return []*pageLayoutT{}, nil
	}

	now := DateTime{time.Now()}
	vlist, ftype, idx := "VERTICAL_LIST", "FIELDS", "0"
	col, row, span := 0.0, 0.0, 12.0
	layouts := make([]*pageLayoutT, 0, len(objs))

	for _, o := range objs {
		layoutID := synthFieldID(o.ID, "record-page-layout")
		tabID := synthFieldID(o.ID, "record-page-tab-home")
		objID := UUID(o.ID)
		icon := "IconLayoutGrid"

		widget := &pageLayoutWidgetT{
			Id: synthFieldID(o.ID, "record-page-widget-fields"), Title: "Fields", Type: "FIELDS",
			IsActive: true, CreatedAt: now, UpdatedAt: now, PageLayoutTabId: tabID,
			GridPosition:  widgetGridPositionT{Column: &col, Row: &row, ColumnSpan: &span, RowSpan: &span},
			Position:      &widgetPositionT{vertical: &pageLayoutWidgetVerticalListPositionT{LayoutMode: &vlist, Index: &idx}},
			Configuration: &widgetConfigurationT{fields: &fieldsConfigurationT{ConfigurationType: &ftype}},
		}
		tab := &pageLayoutTabT{
			Id: tabID, Title: "Home", Icon: &icon, Position: 0, LayoutMode: vlist,
			Widgets: []*pageLayoutWidgetT{widget}, PageLayoutId: layoutID, IsActive: true,
			CreatedAt: now, UpdatedAt: now,
		}
		layouts = append(layouts, &pageLayoutT{
			Id: layoutID, Name: o.LabelSingular + " Record Page", Type: "RECORD_PAGE",
			ObjectMetadataId: &objID, CreatedAt: now, UpdatedAt: now,
			Tabs: []*pageLayoutTabT{tab},
		})
	}
	return layouts, nil
}
