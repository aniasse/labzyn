package twentyapi

import (
	"time"

	"github.com/aniasse/labzyn/server/internal/metadata"
)

// Twenty's frontend assumes certain standard objects always exist (workspaceMember
// is referenced everywhere for actors/avatars). We synthesize them so the app
// renders; their records resolve empty for now.

const (
	workspaceMemberObjectID = "20202020-1c25-4d02-bf25-6aeccf7ea419"
	personObjectID          = "20202020-e674-48e5-a542-72570eee7213"
	companyObjectID         = "20202020-b374-4779-a561-80086cb2e17f"
)

func standardObjects() []objectWithFields {
	now := time.Now()

	wm := stdObject(workspaceMemberObjectID, "workspaceMember", "workspaceMembers", "Workspace Member", "Workspace Members", "IconUserCircle", now,
		stdField(workspaceMemberObjectID, "name", "Name", "FULL_NAME", now),
		stdField(workspaceMemberObjectID, "colorScheme", "Color Scheme", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "avatarUrl", "Avatar Url", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "locale", "Language", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "userEmail", "User Email", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "timeZone", "Time Zone", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "dateFormat", "Date Format", metadata.FieldText, now),
		stdField(workspaceMemberObjectID, "timeFormat", "Time Format", metadata.FieldText, now),
	)
	person := stdObject(personObjectID, "person", "people", "Person", "People", "IconUser", now,
		stdField(personObjectID, "name", "Name", "FULL_NAME", now),
		stdField(personObjectID, "city", "City", metadata.FieldText, now),
		stdField(personObjectID, "jobTitle", "Job Title", metadata.FieldText, now),
		stdField(personObjectID, "avatarUrl", "Avatar Url", metadata.FieldText, now),
	)
	company := stdObject(companyObjectID, "company", "companies", "Company", "Companies", "IconBuildingSkyscraper", now,
		stdField(companyObjectID, "name", "Name", metadata.FieldText, now),
		stdField(companyObjectID, "employees", "Employees", metadata.FieldNumber, now),
	)
	return []objectWithFields{wm, person, company}
}

func stdObject(id, ns, np, ls, lp, icon string, now time.Time, fields ...metadata.Field) objectWithFields {
	return objectWithFields{
		Object: metadata.Object{
			ID: id, NameSingular: ns, NamePlural: np, LabelSingular: ls, LabelPlural: lp,
			Labels: []byte("{}"), Icon: icon, IsSystem: true, IsActive: true, Storage: "jsonb",
			CreatedAt: now, UpdatedAt: now,
		},
		Fields: fields,
	}
}

type objectWithFields struct {
	Object metadata.Object
	Fields []metadata.Field
}

func stdField(objectID, name, label string, t metadata.FieldType, now time.Time) metadata.Field {
	return metadata.Field{
		ID:         string(synthFieldID(objectID, name)),
		ObjectID:   objectID,
		Name:       name,
		Type:       t,
		Label:      label,
		IsSystem:   true,
		IsActive:   true,
		IsNullable: true,
		CreatedAt:  now,
		UpdatedAt:  now,
	}
}
