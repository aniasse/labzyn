package twentyapi

import "context"

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

type widgetPositionT struct{}

func (*widgetPositionT) ToPageLayoutWidgetGridPosition() (*pageLayoutWidgetGridPositionT, bool) {
	return nil, false
}

func (*widgetPositionT) ToPageLayoutWidgetVerticalListPosition() (*pageLayoutWidgetVerticalListPositionT, bool) {
	return nil, false
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

type widgetConfigurationT struct{}

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

func (*widgetConfigurationT) ToFieldsConfiguration() (*fieldsConfigurationT, bool) { return nil, false }

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

func (r *MetadataResolver) GetPageLayouts(_ context.Context, _ struct{ PageLayoutType *string }) ([]*pageLayoutT, error) {
	return []*pageLayoutT{}, nil
}
